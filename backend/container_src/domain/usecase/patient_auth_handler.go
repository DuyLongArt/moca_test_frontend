package usecase

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"regexp"
	"server/infas"
	"strings"
)

var patientPhonePattern = regexp.MustCompile(`^0\d{9}$`)

func normalizePatientPhone(value string) string {
	return strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(strings.TrimSpace(value))
}

func decodePatientAuthRequest(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(target); err != nil {
		writeStandardJSON(w, http.StatusBadRequest, StandardResponse{
			Status:  "ERROR",
			Code:    http.StatusBadRequest,
			Message: "Cannot parse JSON payload",
		})
		return false
	}
	return true
}

func PatientAuthStatusHandler(db *infas.DB_connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var request struct {
			PhoneNumber string `json:"phone_number"`
			Phone       string `json:"phone"`
		}
		if !decodePatientAuthRequest(w, r, &request) {
			return
		}
		phone := normalizePatientPhone(request.PhoneNumber)
		if phone == "" {
			phone = normalizePatientPhone(request.Phone)
		}
		if !patientPhonePattern.MatchString(phone) {
			writeStandardJSON(w, http.StatusBadRequest, StandardResponse{
				Status:  "ERROR",
				Code:    http.StatusBadRequest,
				Message: "A valid phone number is required",
			})
			return
		}
		info, found, err := db.GetPatientAuthInfo(r.Context(), phone)
		if err != nil {
			writeStandardJSON(w, http.StatusInternalServerError, StandardResponse{
				Status:  "ERROR",
				Code:    http.StatusInternalServerError,
				Message: "Could not check patient account",
			})
			return
		}
		writeStandardJSON(w, http.StatusOK, StandardResponse{
			Status:  "OK",
			Code:    http.StatusOK,
			Message: "OK",
			Data: map[string]any{
				"exists":       found,
				"has_password": found && info.PasswordHash != "",
			},
		})
	}
}

func PatientSignupHandler(db *infas.DB_connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var request struct {
			PhoneNumber string `json:"phone_number"`
			Phone       string `json:"phone"`
			Email       string `json:"email"`
			FullName    string `json:"full_name"`
			Password    string `json:"password"`
		}
		if !decodePatientAuthRequest(w, r, &request) {
			return
		}
		phone := normalizePatientPhone(request.PhoneNumber)
		if phone == "" {
			phone = normalizePatientPhone(request.Phone)
		}
		email := strings.ToLower(strings.TrimSpace(request.Email))
		fullName := strings.TrimSpace(request.FullName)
		if !patientPhonePattern.MatchString(phone) {
			writeStandardJSON(w, http.StatusBadRequest, StandardResponse{
				Status:  "ERROR",
				Code:    http.StatusBadRequest,
				Message: "A valid phone number is required",
			})
			return
		}
		if _, err := mail.ParseAddress(email); err != nil {
			writeStandardJSON(w, http.StatusBadRequest, StandardResponse{
				Status:  "ERROR",
				Code:    http.StatusBadRequest,
				Message: "A valid email is required",
			})
			return
		}
		if fullName == "" || len(request.Password) < 8 {
			writeStandardJSON(w, http.StatusBadRequest, StandardResponse{
				Status:  "ERROR",
				Code:    http.StatusBadRequest,
				Message: "Full name and a password of at least 8 characters are required",
			})
			return
		}

		id, _, err := db.CreatePatientPassword(r.Context(), phone, email, fullName, request.Password)
		if errors.Is(err, infas.ErrPatientPasswordExists) {
			writeStandardJSON(w, http.StatusConflict, StandardResponse{
				Status:  "ERROR",
				Code:    http.StatusConflict,
				Message: "This phone number already has a password. Sign in or reset it.",
			})
			return
		}
		if err != nil {
			writeStandardJSON(w, http.StatusBadRequest, StandardResponse{
				Status:  "ERROR",
				Code:    http.StatusBadRequest,
				Message: "Could not create patient password",
			})
			return
		}

		token, err := GenerateToken(phone, "Patient")
		if err != nil {
			writeStandardJSON(w, http.StatusInternalServerError, StandardResponse{
				Status:  "ERROR",
				Code:    http.StatusInternalServerError,
				Message: "Could not generate token",
			})
			return
		}
		writeStandardJSON(w, http.StatusOK, StandardResponse{
			Status:  "OK",
			Code:    http.StatusOK,
			Message: "Patient password created",
			Data: map[string]any{
				"token":        token,
				"id":           id,
				"email":        email,
				"full_name":    fullName,
				"phone_number": phone,
			},
		})
	}
}
