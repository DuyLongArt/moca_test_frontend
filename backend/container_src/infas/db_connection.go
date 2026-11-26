package infas

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
)

type DB_connection struct {
	DB *sql.DB
}

func (app *DB_connection) OpenDB() error {
	db, err := sql.Open("d1", "moca")
	if err != nil {
		return fmt.Errorf("open D1 database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping D1 database: %w", err)
	}
	app.DB = db
	return nil
}

func (app *DB_connection) InitDB() error {
	ctx := context.Background()
	path := strings.TrimSpace(os.Getenv("SCHEMA_PATH"))
	if path == "" {
		path = "/app/schema.sql"
	}
	schema, err := os.ReadFile(path)
	if err != nil {
		schema, err = os.ReadFile("infas/schema.sql")
		if err != nil {
			return fmt.Errorf("read schema: %w", err)
		}
	}
	for _, statement := range strings.Split(string(schema), ";") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := app.DB.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("apply schema: %w", err)
		}
	}

	migrations := []string{
		"ALTER TABLE patients ADD COLUMN email TEXT",
		"ALTER TABLE patients ADD COLUMN password_hash TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE patients ADD COLUMN status TEXT NOT NULL DEFAULT 'Waiting for graded'",
		"ALTER TABLE patients ADD COLUMN doctor_id INTEGER",
		"ALTER TABLE password_reset_token ADD COLUMN role TEXT NOT NULL DEFAULT 'doctor'",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_patients_email ON patients(email)",
		"INSERT OR IGNORE INTO patients (id, full_name, phone_number, password_hash, created_at) SELECT id, full_name, phone_number, COALESCE(password_hash, ''), created_at FROM patient",
		"INSERT OR IGNORE INTO doctors (id, email, password_hash, full_name, specialty, license_number, created_at) SELECT id, email, password_hash, full_name, NULL, NULL, created_at FROM doctor",
	}
	for _, migration := range migrations {
		_, _ = app.DB.ExecContext(ctx, migration)
	}
	return nil
}
