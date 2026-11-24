package usecase

type DoctorLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Jwt      string `json:"jwt"`
}
