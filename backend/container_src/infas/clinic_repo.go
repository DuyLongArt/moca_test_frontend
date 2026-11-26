package infas

import (
	"context"
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type PatientListItem struct {
	ID          int64  `json:"id"`
	DoctorID    int64  `json:"doctor_id"`
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
	Status      string `json:"status"`
	Score       *int64 `json:"score"`
}

type DoctorListItem struct {
	ID        int64  `json:"id,omitempty"`
	FullName  string `json:"full_name"`
	Specialty string `json:"specialty"`
}

type GradeResult struct {
	TestID     int64 `json:"test_id"`
	AddedScore int64 `json:"added_score"`
	Score      int64 `json:"score"`
}

func (app *DB_connection) ListDoctorPatients(ctx context.Context, doctorID int64) ([]PatientListItem, error) {
	rows, err := app.DB.QueryContext(
		ctx,
		`SELECT p.id, p.doctor_id, p.full_name, p.phone_number, p.status, t.score
		 FROM patients p
		 LEFT JOIN tests t ON t.patient_id = p.id
		 WHERE p.doctor_id = ?
		 ORDER BY p.id`,
		doctorID,
	)
	if err != nil {
		return nil, fmt.Errorf("list doctor patients: %w", err)
	}
	defer rows.Close()

	patients := []PatientListItem{}
	for rows.Next() {
		var patient PatientListItem
		var score sql.NullInt64
		if err := rows.Scan(
			&patient.ID,
			&patient.DoctorID,
			&patient.FullName,
			&patient.PhoneNumber,
			&patient.Status,
			&score,
		); err != nil {
			return nil, fmt.Errorf("scan doctor patient: %w", err)
		}
		if score.Valid {
			value := score.Int64
			patient.Score = &value
		}
		patients = append(patients, patient)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate doctor patients: %w", err)
	}
	return patients, nil
}

func (app *DB_connection) AssignPatientDoctor(ctx context.Context, patientID, doctorID int64) (bool, error) {
	result, err := app.DB.ExecContext(
		ctx,
		"UPDATE patients SET doctor_id = ? WHERE id = ?",
		doctorID,
		patientID,
	)
	if err != nil {
		return false, fmt.Errorf("assign patient doctor: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("check assigned patient: %w", err)
	}
	return affected > 0, nil
}

func (app *DB_connection) GradeTest(ctx context.Context, testID, addedScore int64) (*GradeResult, bool, error) {
	result, err := app.DB.ExecContext(
		ctx,
		"UPDATE tests SET score = score + ? WHERE id = ?",
		addedScore,
		testID,
	)
	if err != nil {
		return nil, false, fmt.Errorf("grade test: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, false, fmt.Errorf("check graded test: %w", err)
	}
	if affected == 0 {
		return nil, false, nil
	}

	var score int64
	if err := app.DB.QueryRowContext(ctx, "SELECT score FROM tests WHERE id = ?", testID).Scan(&score); err != nil {
		return nil, false, fmt.Errorf("read graded test: %w", err)
	}
	return &GradeResult{TestID: testID, AddedScore: addedScore, Score: score}, true, nil
}

func (app *DB_connection) UpdatePatientScore(ctx context.Context, patientID, score int64) (bool, error) {
	result, err := app.DB.ExecContext(
		ctx,
		`INSERT INTO tests (patient_id, score)
		 SELECT ?, ?
		 WHERE NOT EXISTS (SELECT 1 FROM tests WHERE patient_id = ?)`,
		patientID,
		score,
		patientID,
	)
	if err != nil {
		return false, fmt.Errorf("update patient score: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("check updated patient score: %w", err)
	}
	return affected > 0, nil
}

func (app *DB_connection) ListDoctors(ctx context.Context) ([]DoctorListItem, error) {
	rows, err := app.DB.QueryContext(ctx, "SELECT id, full_name, specialty FROM doctors ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("list doctors: %w", err)
	}
	defer rows.Close()

	doctors := []DoctorListItem{}
	for rows.Next() {
		var doctor DoctorListItem
		var specialty sql.NullString
		if err := rows.Scan(&doctor.ID, &doctor.FullName, &specialty); err != nil {
			return nil, fmt.Errorf("scan doctor: %w", err)
		}
		doctor.Specialty = specialty.String
		doctors = append(doctors, doctor)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate doctors: %w", err)
	}
	return doctors, nil
}

func (app *DB_connection) CreateDoctorAccount(ctx context.Context, email, password, fullName string) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hash doctor password: %w", err)
	}
	result, err := app.DB.ExecContext(
		ctx,
		"INSERT INTO doctors (email, password_hash, full_name) VALUES (?, ?, ?)",
		email,
		string(hash),
		fullName,
	)
	if err != nil {
		return 0, fmt.Errorf("create doctor account: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read doctor id: %w", err)
	}
	return id, nil
}
