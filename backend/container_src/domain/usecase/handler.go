package usecase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/infas"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	// "server/infas"
)

func PatientHandler(db *infas.DB_connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		var request struct {
			PhoneNumber  string `json:"phone_number"`
			PhoneSnake   string `json:"phoneNumber"`
			Phone        string `json:"phone"`
			Password     string `json:"password"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&request); err != nil {
			writeStandardJSON(w, http.StatusBadRequest, StandardResponse{
				Status:  "ERROR",
				Code:    http.StatusBadRequest,
				Message: "Cannot parse JSON payload",
			})
			return
		}
		phone := strings.TrimSpace(request.PhoneNumber)
		if phone == "" {
			phone = strings.TrimSpace(request.PhoneSnake)
		}
		if phone == "" {
			phone = strings.TrimSpace(request.Phone)
		}
		if phone == "" || len(request.Password) < 8 {
			writeStandardJSON(w, http.StatusBadRequest, StandardResponse{
				Status:  "ERROR",
				Code:    http.StatusBadRequest,
				Message: "A valid phone and password of at least 8 characters are required",
			})
			return
		}

		info, found, err := db.GetPatientAuthInfo(r.Context(), phone)
		if err != nil {
			writeStandardJSON(w, http.StatusInternalServerError, StandardResponse{
				Status:  "ERROR",
				Code:    http.StatusInternalServerError,
				Message: "Could not sign in",
			})
			return
		}
		if !found || info.PasswordHash == "" {
			writeStandardJSON(w, http.StatusUnauthorized, StandardResponse{
				Status:  "ERROR",
				Code:    http.StatusUnauthorized,
				Message: "Create a password before signing in",
			})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(info.PasswordHash), []byte(request.Password)); err != nil {
			writeStandardJSON(w, http.StatusUnauthorized, StandardResponse{
				Status:  "ERROR",
				Code:    http.StatusUnauthorized,
				Message: "Invalid phone or password",
			})
			return
		}

		token, err := GenerateToken(info.PhoneNumber, "Patient")
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
			Message: "Login succesfully",
			Data: map[string]any{
				"token":        token,
				"id":           info.ID,
				"email":        info.Email,
				"full_name":    info.FullName,
				"phone_number": info.PhoneNumber,
			},
		})
	}
}

func DoctorHandler(db_connector *infas.DB_connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		var request DoctorLoginRequest
		var response StandardResponse

		if errorJson := json.NewDecoder(r.Body).Decode(&request); errorJson != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(StandardResponse{
				Status:  "ERROR",
				Code:    http.StatusBadRequest,
				Message: "Cannot parse JSON payload",
			})
			return
		}
		if request.Jwt == "" {

			email, hash_password, errorDb := db_connector.GetEmailPassFromDoctor(r.Context(), request.Email)
			if hash_password != request.Password || errorDb != nil {

				if errorDb != nil {

					http.Error(w, "Invalid email or password", http.StatusUnauthorized)
					return
				}

				// 2. Verify incoming plaintext password against the stored bcrypt hash
				if err := bcrypt.CompareHashAndPassword([]byte(hash_password), []byte(request.Password)); err != nil {
					// Password does not match
					http.Error(w, "Update user account", http.StatusUnauthorized)
					db_connector.UpsertDoctorAccount(r.Context(), request.Email, request.Password)
					return
				}

			} else {

			}
			if len(hash_password) > 0 {
				jwt, error := GenerateToken(email, "Doctor")
				if error != nil {
					http.Error(w, "Can not generate token", http.StatusInternalServerError)
				}

				response.Data = map[string]any{
					"token": jwt,
				}
				if error != nil {
					http.Error(w, "The request is not encodeded", http.StatusBadRequest)
					return
				}
			} else {

			}

			response.Status = "OK"
			response.Code = 200
			response.Message = "Login succesfully"
			json.NewEncoder(w).Encode(response)
		} else {

			claims := jwt.MapClaims{}
			parsedToken, err := jwt.ParseWithClaims(request.Jwt, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return LoadJWTSecret(), nil
			})

			if err != nil || !parsedToken.Valid {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Verify token identity matches requested email
			if tokenEmail, ok := claims["user_id"].(string); !ok || tokenEmail != request.Email {
				http.Error(w, "Token does not match requested email", http.StatusForbidden)
				return
			}

			response.Status = "OK"
			response.Code = http.StatusOK
			response.Message = "Token is valid"
			response.Data = map[string]any{
				"email": claims["user_id"],
				"role":  claims["role"],
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}
	}
}
