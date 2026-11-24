package usecase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/infas"

	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	// "server/infas"
)

func PatientHandler(db_connector *infas.DB_connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		var request PatientLoginRequest
		var response StandardResponse
		error := json.NewDecoder(r.Body).Decode(&request)
		email, hash_password, error := db_connector.GetEmailPassFromDoctor(r.Context(), request.PhoneNumber)

		if error != nil {
			http.Error(w, "Invalid credential, sign up new account", http.StatusUnauthorized)
			db_connector.AddPatientAccount(r.Context(), request.PhoneNumber)
			return
		}

		if len(hash_password) > 0 {
			jwt, error := GenerateToken(email, "Patient")
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
				jwt, error := GenerateToken(email, "Patient")
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
