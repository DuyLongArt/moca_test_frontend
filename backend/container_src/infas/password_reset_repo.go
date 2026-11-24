package infas

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const passwordResetTTL = 30 * time.Minute

func (app *DB_connection) CreatePasswordResetToken(ctx context.Context, email string) (string, bool, error) {
	var doctorID int64
	var storedEmail string
	err := app.DB.QueryRowContext(
		ctx,
		"SELECT id, email FROM doctor WHERE lower(email) = ? LIMIT 1",
		email,
	).Scan(&doctorID, &storedEmail)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("find doctor for password reset: %w", err)
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", false, fmt.Errorf("generate password reset token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	digest := sha256.Sum256([]byte(token))

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", false, fmt.Errorf("start password reset transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM password_reset_token WHERE doctor_id = ?", doctorID); err != nil {
		return "", false, fmt.Errorf("clear old password reset tokens: %w", err)
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO password_reset_token (doctor_id, email, token_hash, expires_at) VALUES (?, ?, ?, ?)",
		doctorID,
		storedEmail,
		digest[:],
		time.Now().Add(passwordResetTTL).Unix(),
	)
	if err != nil {
		return "", false, fmt.Errorf("store password reset token: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return "", false, fmt.Errorf("commit password reset token: %w", err)
	}

	return token, true, nil
}

func (app *DB_connection) ResetDoctorPassword(ctx context.Context, token, newPassword string) (bool, error) {
	digest := sha256.Sum256([]byte(token))

	var tokenID int64
	var doctorID int64
	err := app.DB.QueryRowContext(
		ctx,
		`SELECT id, doctor_id FROM password_reset_token
		 WHERE token_hash = ? AND used_at IS NULL AND expires_at > ?
		 LIMIT 1`,
		digest[:],
		time.Now().Unix(),
	).Scan(&tokenID, &doctorID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("find password reset token: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return false, fmt.Errorf("hash new password: %w", err)
	}

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("start password reset transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(
		ctx,
		"UPDATE doctor SET password_hash = ? WHERE id = ?",
		string(hash),
		doctorID,
	)
	if err != nil {
		return false, fmt.Errorf("update doctor password: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("check doctor password update: %w", err)
	}
	if affected == 0 {
		return false, nil
	}

	if _, err := tx.ExecContext(
		ctx,
		"UPDATE password_reset_token SET used_at = ? WHERE id = ?",
		time.Now().Unix(),
		tokenID,
	); err != nil {
		return false, fmt.Errorf("consume password reset token: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit password reset: %w", err)
	}

	return true, nil
}
