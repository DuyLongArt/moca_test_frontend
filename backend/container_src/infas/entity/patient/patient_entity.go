package entity

type Patient struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	FullName     string `json:"full_name"`
	PasswordHash string `json:"-"`
	PhoneNumber  string `json:"phone_number"`
}

func (p *Patient) GetEmail() string {
	return p.Email
}

func (p *Patient) GetPasswordHash() string {
	return p.PasswordHash
}

func (p *Patient) GetPhoneNumber() string {
	return p.PhoneNumber
}

func (p *Patient) GetName() string {
	return p.FullName
}
