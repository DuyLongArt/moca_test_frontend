package usecase

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"server/infas"
)

func writeStandardJSON(w http.ResponseWriter, status int, body StandardResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeClinicError(w http.ResponseWriter, status int, message string) {
	writeStandardJSON(w, status, StandardResponse{
		Status:  "ERROR",
		Code:    status,
		Message: message,
	})
}

func decodeClinicJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := decoder.Decode(target); err != nil {
		writeClinicError(w, http.StatusBadRequest, "Cannot parse JSON payload")
		return false
	}
	return true
}

func parsePositiveQueryInt64(r *http.Request, key string) (int64, bool) {
	value, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get(key)), 10, 64)
	if err != nil || value <= 0 {
		return 0, false
	}
	return value, true
}

func DoctorPatientListHandler(db *infas.DB_connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeClinicError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		doctorID, ok := parsePositiveQueryInt64(r, "id")
		if !ok {
			writeClinicError(w, http.StatusBadRequest, "A valid doctor id is required")
			return
		}
		patients, err := db.ListDoctorPatients(r.Context(), doctorID)
		if err != nil {
			writeClinicError(w, http.StatusInternalServerError, "Could not list patients")
			return
		}
		writeStandardJSON(w, http.StatusOK, StandardResponse{
			Status:  "OK",
			Code:    http.StatusOK,
			Message: "Patients list",
			Data:    patients,
		})
	}
}

func AssignPatientDoctorHandler(db *infas.DB_connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeClinicError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		patientID, ok := parsePositiveQueryInt64(r, "patient_id")
		if !ok {
			writeClinicError(w, http.StatusBadRequest, "A valid patient_id is required")
			return
		}
		doctorID, ok := parsePositiveQueryInt64(r, "doctor_id")
		if !ok {
			writeClinicError(w, http.StatusBadRequest, "A valid doctor_id is required")
			return
		}
		assigned, err := db.AssignPatientDoctor(r.Context(), patientID, doctorID)
		if err != nil {
			writeClinicError(w, http.StatusInternalServerError, "Could not assign patient")
			return
		}
		if !assigned {
			writeClinicError(w, http.StatusNotFound, "Patient not found")
			return
		}
		writeStandardJSON(w, http.StatusOK, StandardResponse{
			Status:  "OK",
			Code:    http.StatusOK,
			Message: "Patient assigned to doctor",
			Data: map[string]any{
				"patient_id": patientID,
				"doctor_id":  doctorID,
			},
		})
	}
}

func GradeTestHandler(db *infas.DB_connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeClinicError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		var request struct {
			TestID int64 `json:"test_id"`
			Score  int64 `json:"score"`
		}
		if !decodeClinicJSON(w, r, &request) {
			return
		}
		if request.TestID <= 0 || request.Score < 1 || request.Score > 30 {
			writeClinicError(w, http.StatusBadRequest, "A valid test_id and score are required")
			return
		}
		result, graded, err := db.GradeTest(r.Context(), request.TestID, request.Score)
		if err != nil {
			writeClinicError(w, http.StatusInternalServerError, "Could not grade test")
			return
		}
		if !graded {
			writeClinicError(w, http.StatusNotFound, "Test not found")
			return
		}
		writeStandardJSON(w, http.StatusOK, StandardResponse{
			Status:  "OK",
			Code:    http.StatusOK,
			Message: "Test graded",
			Data:    result,
		})
	}
}

func UpdatePatientScoreHandler(db *infas.DB_connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeClinicError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		var request struct {
			PatientID int64 `json:"patient_id"`
			Score     int64 `json:"score"`
		}
		if !decodeClinicJSON(w, r, &request) {
			return
		}
		if request.PatientID <= 0 || request.Score < 1 || request.Score > 30 {
			writeClinicError(w, http.StatusBadRequest, "A valid patient_id and score are required")
			return
		}
		if _, err := db.UpdatePatientScore(r.Context(), request.PatientID, request.Score); err != nil {
			writeClinicError(w, http.StatusInternalServerError, "Could not update patient score")
			return
		}
		writeStandardJSON(w, http.StatusOK, StandardResponse{
			Status:  "OK",
			Code:    http.StatusOK,
			Message: "Patient score updated",
			Data: map[string]any{
				"patient_id": request.PatientID,
				"score":      request.Score,
			},
		})
	}
}

func DoctorsListHandler(db *infas.DB_connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeClinicError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		doctors, err := db.ListDoctors(r.Context())
		if err != nil {
			writeClinicError(w, http.StatusInternalServerError, "Could not list doctors")
			return
		}
		writeStandardJSON(w, http.StatusOK, StandardResponse{
			Status:  "OK",
			Code:    http.StatusOK,
			Message: "Doctors list",
			Data:    doctors,
		})
	}
}

func DoctorSignupHandler(db *infas.DB_connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeClinicError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		var request struct {
			Email         string `json:"email"`
			Password      string `json:"password"`
			FullName      string `json:"fullName"`
			FullNameSnake string `json:"full_name"`
		}
		if !decodeClinicJSON(w, r, &request) {
			return
		}
		email, ok := normalizeResetEmail(request.Email)
		if !ok || len(request.Password) < 8 {
			writeClinicError(w, http.StatusBadRequest, "A valid email and password of at least 8 characters are required")
			return
		}
		fullName := strings.TrimSpace(request.FullName)
		if fullName == "" {
			fullName = strings.TrimSpace(request.FullNameSnake)
		}
		if fullName == "" {
			writeClinicError(w, http.StatusBadRequest, "fullName is required")
			return
		}
		id, err := db.CreateDoctorAccount(r.Context(), email, request.Password, fullName)
		if err != nil {
			writeClinicError(w, http.StatusConflict, "Could not create doctor account")
			return
		}
		writeStandardJSON(w, http.StatusCreated, StandardResponse{
			Status:  "OK",
			Code:    http.StatusCreated,
			Message: "Doctor account created",
			Data: map[string]any{
				"id":    id,
				"email": email,
			},
		})
	}
}
