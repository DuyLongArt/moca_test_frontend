package domain

import (
	"context"
	auth "server/domain/usecase"
	"server/infas"
	entity "server/infas/entity/doctor"
)

func authenticate(doctor_request *auth.DoctorLoginRequest) *entity.Doctor {
	db := &infas.DB_connection{}

	var doctor_entity, error = db.AuthenticateDoctor(context.Background(), doctor_request.Email, doctor_request.Password)
	if error != nil {
		return &entity.Doctor{}
	}

	return doctor_entity

}
