package infas

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	infas "server/infas/entity/patient"

	"golang.org/x/crypto/bcrypt"
	// "server/domain/patient" // update with your actual package path
	// "server/infas/entity/patient" // update with your actual package path
)

func (app *DB_connection) AuthenticatePatient(ctx context.Context, email, password string) (*infas.Patient, error) {
	query := `
		SELECT id, email, full_name, password_hash, phone_number
		FROM patient
		WHERE email = $1
		LIMIT 1
	`

	var patient infas.Patient

	// 1. Query the database for the user by email
	err := app.DB.QueryRowContext(ctx, query, email).Scan(
		&patient.ID,
		&patient.Email,
		&patient.FullName,
		&patient.PasswordHash,
		&patient.PhoneNumber,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Email not found (use generic error message for security)
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("failed to query patient: %w", err)
	}

	// 2. Verify the hashed password against the plain-text password input
	err = bcrypt.CompareHashAndPassword([]byte(patient.PasswordHash), []byte(password))
	if err != nil {
		// Password does not match
		return nil, errors.New("invalid credentials")
	}

	return &patient, nil
}

func (app *DB_connection) GetPhoneAndPasswordFromPatient(ctx context.Context, requestPhoneNumber string) (string, string, error) {
	query := "SELECT phone_number, hash_password FROM patient WHERE phone_number =$1"
	var phone_number string
	var hash_password string
	err := app.DB.QueryRowContext(ctx, query, requestPhoneNumber).Scan(&phone_number, &hash_password)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}
	return phone_number, hash_password, nil
}

func (app *DB_connection) AddPatientAccount(ctx context.Context, phoneNumber string) (int64, error) {
	query := `
		INSERT INTO patient (
			phone_number, 
			password_hash, 
			full_name, 
			created_at
		) VALUES (?, ?, ?, ?)
	`
	result, err := app.DB.ExecContext(ctx, query, phoneNumber, "", "", "", "")

	if err != nil {
		return 0, err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastID, nil

}
