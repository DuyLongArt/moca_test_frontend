package usecase

import (
	"context"
	"fmt"
	"server/infas"
)

type User interface {
	GetPhoneNumber() string
	GetEmail() string
	GetPasswordHash() string
	GetName() string
}
type Authenticatable interface {
	Authenticate(ctx context.Context, db *infas.DB_connection) (User, error)
}

// Authenticate Patient
func (req *PatientLoginRequest) Authenticate(ctx context.Context, db *infas.DB_connection) (User, error) {
	patientEntity, err := db.AuthenticatePatient(ctx, req.PhoneNumber, req.Password)
	if err != nil {
		return nil, fmt.Errorf("patient authentication failed: %w", err)
	}
	return patientEntity, nil
}

// Authenticate Doctor
func (req *DoctorLoginRequest) Authenticate(ctx context.Context, db *infas.DB_connection) (User, error) {
	doctorEntity, err := db.AuthenticateDoctor(ctx, req.Email, req.Password)
	if err != nil {
		return nil, fmt.Errorf("doctor authentication failed: %w", err)
	}
	return doctorEntity, nil
}

// Service xử lý đăng nhập chung cho mọi đối tượng Authenticatable
func Login(ctx context.Context, db *infas.DB_connection, req Authenticatable) (User, error) {
	user, err := req.Authenticate(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("login process failed: %w", err)
	}

	// Các logic xử lý sau khi xác thực thành công (như tạo JWT token, ghi log...)
	// Đều có thể dùng interface User (GetName, GetEmail, v.v.)
	fmt.Printf("User %s (%s) logged in successfully!\n", user.GetName(), user.GetEmail())

	return user, nil
}
