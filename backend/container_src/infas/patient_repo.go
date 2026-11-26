package infas

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	entity "server/infas/entity/patient"
)

var (
	ErrPatientPasswordMissing = errors.New("patient password has not been created")
	ErrPatientPasswordExists  = errors.New("patient password already exists")
)

type PatientAuthInfo struct {
	ID           int64
	Email        string
	FullName     string
	PhoneNumber  string
	PasswordHash string
}

func (app *DB_connection) GetPatientAuthInfo(ctx context.Context, phoneNumber string) (*PatientAuthInfo, bool, error) {
	var patient PatientAuthInfo
	err := app.DB.QueryRowContext(
		ctx,
		`SELECT id, COALESCE(email, ''), full_name, phone_number, COALESCE(password_hash, '')
		 FROM patients WHERE phone_number = ? LIMIT 1`,
		phoneNumber,
	).Scan(
		&patient.ID,
		&patient.Email,
		&patient.FullName,
		&patient.PhoneNumber,
		&patient.PasswordHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("find patient: %w", err)
	}
	return &patient, true, nil
}

func (app *DB_connection) AuthenticatePatient(ctx context.Context, phoneNumber, password string) (*entity.Patient, error) {
	info, found, err := app.GetPatientAuthInfo(ctx, phoneNumber)
	if err != nil {
		return nil, err
	}
	if !found || info.PasswordHash == "" {
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(info.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return &entity.Patient{
		ID:           int(info.ID),
		FullName:     info.FullName,
		PasswordHash: info.PasswordHash,
		PhoneNumber:  info.PhoneNumber,
	}, nil
}

func (app *DB_connection) CreatePatientPassword(
	ctx context.Context,
	phoneNumber,
	email,
	fullName,
	password string,
) (int64, bool, error) {
	phoneNumber = strings.TrimSpace(phoneNumber)
	email = strings.ToLower(strings.TrimSpace(email))
	fullName = strings.TrimSpace(fullName)
	if phoneNumber == "" || email == "" || fullName == "" || len(password) < 8 {
		return 0, false, errors.New("phone, email, full name and password of at least 8 characters are required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, false, fmt.Errorf("hash patient password: %w", err)
	}

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, false, fmt.Errorf("start patient password transaction: %w", err)
	}
	defer tx.Rollback()

	var id int64
	var storedHash string
	err = tx.QueryRowContext(
		ctx,
		"SELECT id, COALESCE(password_hash, '') FROM patients WHERE phone_number = ? LIMIT 1",
		phoneNumber,
	).Scan(&id, &storedHash)
	created := errors.Is(err, sql.ErrNoRows)
	if err != nil && !created {
		return 0, false, fmt.Errorf("find patient: %w", err)
	}

	if created {
		result, err := tx.ExecContext(
			ctx,
			"INSERT INTO patients (phone_number, password_hash, full_name, email) VALUES (?, ?, ?, ?)",
			phoneNumber,
			string(hash),
			fullName,
			email,
		)
		if err != nil {
			return 0, false, fmt.Errorf("create patient account: %w", err)
		}
		id, err = result.LastInsertId()
		if err != nil {
			return 0, false, fmt.Errorf("read patient id: %w", err)
		}
	} else {
		if storedHash != "" {
			return 0, false, ErrPatientPasswordExists
		}
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE patients SET password_hash = ?, full_name = COALESCE(NULLIF(?, ''), full_name), email = ?
			 WHERE id = ?`,
			string(hash),
			fullName,
			email,
			id,
		); err != nil {
			return 0, false, fmt.Errorf("set patient password: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, false, fmt.Errorf("commit patient password: %w", err)
	}
	return id, created, nil
}

func (app *DB_connection) GetPhoneAndPasswordFromPatient(ctx context.Context, requestPhoneNumber string) (string, string, error) {
	info, found, err := app.GetPatientAuthInfo(ctx, requestPhoneNumber)
	if err != nil {
		return "", "", err
	}
	if !found {
		return "", "", errors.New("invalid credentials")
	}
	return info.PhoneNumber, info.PasswordHash, nil
}

func (app *DB_connection) AddPatientAccount(ctx context.Context, phoneNumber string) (int64, error) {
	query := `
		INSERT INTO patients (
			phone_number,
			password_hash,
			full_name
		) VALUES (?, ?, ?)
	`
	result, err := app.DB.ExecContext(ctx, query, phoneNumber, "", "")
	if err != nil {
		return 0, err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastID, nil
}
