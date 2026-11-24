package infas

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	// "server/domain/doctor" // update with your actual package path

	infas "server/infas/entity/doctor" // update with your actual package path
)

func (app *DB_connection) AuthenticateDoctor(ctx context.Context, email, password string) (*infas.Doctor, error) {
	query := `
		SELECT id, email, full_name, password_hash, specialty, license_number
		FROM doctor
		WHERE email = $1
		LIMIT 1
	`

	var doctor infas.Doctor

	// 1. Query the database for the user by email
	err := app.DB.QueryRowContext(ctx, query, email).Scan(
		&doctor.ID,
		&doctor.Email,
		&doctor.FullName,
		&doctor.PasswordHash,
		&doctor.Specialty,
		&doctor.LicenseNumber,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Email not found (use gene ric error message for security)
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("failed to query doctor: %w", err)
	}

	// 2. Verify the hashed password against the plain-text password input
	err = bcrypt.CompareHashAndPassword([]byte(doctor.PasswordHash), []byte(password))
	if err != nil {
		// Password does not match
		return nil, errors.New("invalid credentials")
	}

	return &doctor, nil
}
func (app *DB_connection) GetEmailPassFromDoctor(ctx context.Context, requestEmail string) (string, string, error) {
	query := "SELECT email, password_hash FROM doctor WHERE email = ?"
	var email string
	var hash_password string
	err := app.DB.QueryRowContext(ctx, query, requestEmail).Scan(&email, &hash_password)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}
	return email, hash_password, nil
}
func (app *DB_connection) UpsertDoctorAccount(
	ctx context.Context,
	email string,
	passwordHash string,
) (int64, error) {
	query := `
        INSERT INTO doctor (
            email, 
            password_hash,
			full_name
        ) VALUES (?, ?, '')
        ON CONFLICT (email) DO UPDATE SET
            password_hash  = excluded.password_hash
           
        RETURNING id;
    `

	var doctorID int64
	err := app.DB.QueryRowContext(
		ctx,
		query,
		email,
		passwordHash,
	).Scan(&doctorID)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert doctor: %w", err)
	}

	return doctorID, nil
}
