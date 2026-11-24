package domain

import (
	// "server/domain/doctor"
	"context"
	auth "server/domain/usecase"

	// "server/infas"
	infas "server/infas"
	patient "server/infas/entity/patient"
)

func authenticate(patient_request *auth.PatientLoginRequest) *patient.Patient {
	var patient_email_request = patient_request.PhoneNumber
	var patient_passwordhash_request = patient_request.Password

	var patient_repo = infas.DB_connection{}
	var patient_entity, error = patient_repo.AuthenticatePatient(context.Background(), patient_email_request, patient_passwordhash_request)
	if error != nil {
		return nil
	}

	return patient_entity

}
