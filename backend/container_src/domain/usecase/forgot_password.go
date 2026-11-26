package usecase

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"server/infas"
	"strings"
)

type passwordResetResponse struct {
	Status  string            `json:"status"`
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    map[string]string `json:"data,omitempty"`
}

func normalizeResetEmail(value string) (string, bool) {
	address, err := mail.ParseAddress(strings.TrimSpace(value))
	if err != nil || address.Address == "" {
		return "", false
	}
	return strings.ToLower(address.Address), true
}

func writePasswordResetJSON(w http.ResponseWriter, status int, body passwordResetResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func ForgotPasswordHandler(db *infas.DB_connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writePasswordResetJSON(w, http.StatusMethodNotAllowed, passwordResetResponse{
				Status:  "ERROR",
				Code:    http.StatusMethodNotAllowed,
				Message: "Method not allowed",
			})
			return
		}

		var request struct {
			Email string `json:"email"`
			Role  string `json:"role"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		if err := decoder.Decode(&request); err != nil {
			writePasswordResetJSON(w, http.StatusBadRequest, passwordResetResponse{
				Status:  "ERROR",
				Code:    http.StatusBadRequest,
				Message: "Cannot parse JSON payload",
			})
			return
		}

		email, ok := normalizeResetEmail(request.Email)
		if !ok {
			writePasswordResetJSON(w, http.StatusBadRequest, passwordResetResponse{
				Status:  "ERROR",
				Code:    http.StatusBadRequest,
				Message: "Invalid email address",
			})
			return
		}

		role := strings.ToLower(strings.TrimSpace(request.Role))
		if role != "patient" {
			role = "doctor"
		}
		token, found, err := db.CreatePasswordResetTokenForRole(r.Context(), email, role)
		if err != nil {
			writePasswordResetJSON(w, http.StatusInternalServerError, passwordResetResponse{
				Status:  "ERROR",
				Code:    http.StatusInternalServerError,
				Message: "Could not start password reset",
			})
			return
		}

		response := passwordResetResponse{
			Status:  "OK",
			Code:    http.StatusOK,
			Message: "If the account exists, a reset link has been created",
		}
		if found {
			response.Data = map[string]string{"token": token, "email": email, "role": role}
		}
		writePasswordResetJSON(w, http.StatusOK, response)
	}
}

func ResetPasswordHandler(db *infas.DB_connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writePasswordResetJSON(w, http.StatusMethodNotAllowed, passwordResetResponse{
				Status:  "ERROR",
				Code:    http.StatusMethodNotAllowed,
				Message: "Method not allowed",
			})
			return
		}

		var request struct {
			Token    string `json:"token"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		if err := decoder.Decode(&request); err != nil {
			writePasswordResetJSON(w, http.StatusBadRequest, passwordResetResponse{
				Status:  "ERROR",
				Code:    http.StatusBadRequest,
				Message: "Cannot parse JSON payload",
			})
			return
		}

		if strings.TrimSpace(request.Token) == "" || len(request.Password) < 8 {
			writePasswordResetJSON(w, http.StatusBadRequest, passwordResetResponse{
				Status:  "ERROR",
				Code:    http.StatusBadRequest,
				Message: "A valid token and a password of at least 8 characters are required",
			})
			return
		}

		role := strings.ToLower(strings.TrimSpace(request.Role))
		if role != "patient" {
			role = "doctor"
		}
		reset, err := db.ResetPasswordForRole(r.Context(), strings.TrimSpace(request.Token), request.Password, role)
		if err != nil {
			writePasswordResetJSON(w, http.StatusInternalServerError, passwordResetResponse{
				Status:  "ERROR",
				Code:    http.StatusInternalServerError,
				Message: "Could not reset password",
			})
			return
		}
		if !reset {
			writePasswordResetJSON(w, http.StatusBadRequest, passwordResetResponse{
				Status:  "ERROR",
				Code:    http.StatusBadRequest,
				Message: "Invalid or expired reset token",
			})
			return
		}

		writePasswordResetJSON(w, http.StatusOK, passwordResetResponse{
			Status:  "OK",
			Code:    http.StatusOK,
			Message: "Password has been reset",
		})
	}
}
