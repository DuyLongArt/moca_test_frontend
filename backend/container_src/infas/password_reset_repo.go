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

const (
	passwordResetDoctor  = "doctor"
	passwordResetPatient = "patient"
)

func normalizePasswordResetRole(role string) string {
	if role == passwordResetPatient {
		return passwordResetPatient
	}
	return passwordResetDoctor
}

func passwordResetSubjectTable(role string) string {
	if role == passwordResetPatient {
		return "patients"
	}
	return "doctors"
}

func findPasswordResetSubject(ctx context.Context, db *sql.DB, email, role string) (int64, string, error) {
	var id int64
	var storedEmail string
	query := "SELECT id, email FROM doctors WHERE lower(email) = ? LIMIT 1"
	if role == passwordResetPatient {
		query = "SELECT id, COALESCE(email, '') FROM patients WHERE lower(COALESCE(email, '')) = ? LIMIT 1"
	}
	err := db.QueryRowContext(ctx, query, email).Scan(&id, &storedEmail)
	return id, storedEmail, err
}

func (app *DB_connection) CreatePasswordResetToken(ctx context.Context, email string) (string, bool, error) {
	return app.CreatePasswordResetTokenForRole(ctx, email, passwordResetDoctor)
}

func (app *DB_connection) CreatePasswordResetTokenForRole(ctx context.Context, email, role string) (string, bool, error) {
	role = normalizePasswordResetRole(role)
	subjectID, storedEmail, err := findPasswordResetSubject(ctx, app.DB, email, role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("find %s for password reset: %w", role, err)
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

	if _, err := tx.ExecContext(
		ctx,
		"DELETE FROM password_reset_token WHERE role = ? AND doctor_id = ?",
		role,
		subjectID,
	); err != nil {
		return "", false, fmt.Errorf("clear old password reset tokens: %w", err)
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO password_reset_token (doctor_id, email, token_hash, expires_at, role) VALUES (?, ?, ?, ?, ?)",
		subjectID,
		storedEmail,
		digest[:],
		time.Now().Add(passwordResetTTL).Unix(),
		role,
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
	return app.ResetPasswordForRole(ctx, token, newPassword, passwordResetDoctor)
}

func (app *DB_connection) ResetPasswordForRole(ctx context.Context, token, newPassword, role string) (bool, error) {
	role = normalizePasswordResetRole(role)
	digest := sha256.Sum256([]byte(token))

	var tokenID int64
	var subjectID int64
	err := app.DB.QueryRowContext(
		ctx,
		`SELECT id, doctor_id FROM password_reset_token
		 WHERE token_hash = ? AND role = ? AND used_at IS NULL AND expires_at > ?
		 LIMIT 1`,
		digest[:],
		role,
		time.Now().Unix(),
	).Scan(&tokenID, &subjectID)
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

	table := passwordResetSubjectTable(role)
	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("start password reset transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(
		ctx,
		fmt.Sprintf("UPDATE %s SET password_hash = ? WHERE id = ?", table),
		string(hash),
		subjectID,
	)
	if err != nil {
		return false, fmt.Errorf("update %s password: %w", role, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("check %s password update: %w", role, err)
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
