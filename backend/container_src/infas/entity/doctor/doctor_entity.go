package entity

type Doctor struct {
	ID            int    `json:"id"`
	Email         string `json:"email" validate:"required,email"`
	PasswordHash  string `json:"-"` // "-" = không bao giờ trả ra JSON
	FullName      string `json:"full_name" validate:"required"`
	Specialty     string `json:"specialty,omitempty"`
	PhoneNumber   string `json:"phone_number,omitempty"`
	LicenseNumber string `json:"license_number,omitempty"`
}

func (p *Doctor) GetEmail() string {
	return p.Email
}

func (p *Doctor) GetPasswordHash() string {
	return p.PasswordHash
}

func (p *Doctor) GetPhoneNumber() string {
	return p.PhoneNumber
}

func (p *Doctor) GetName() string {
	return p.FullName
}
