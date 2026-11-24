package usecase

type PatientLoginRequest struct {
	PhoneNumber string `json:"email"`
	Password    string `json:"password"`
}
