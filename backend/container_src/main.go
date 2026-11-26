package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	// "server/domain/usecase"
	"server/domain/usecase"
	infas "server/infas"

	"syscall"
	"time"
	// "server/domain/usecase"
)

func handler(w http.ResponseWriter, r *http.Request) {
	message := os.Getenv("MESSAGE")
	instanceID := os.Getenv("INSTANCE_ID")
	fmt.Fprintf(w, "Hi, I'm a container and this is my message: \"%s\", my instance ID is: %s", message, instanceID)
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Exiting with code 1 for the /error demonstration")
	os.Exit(1)
}

func newRouter(app *infas.DB_connection) http.Handler {

	router := http.NewServeMux()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	router.HandleFunc("/", handler)
	router.HandleFunc("/container", handler)
	router.HandleFunc("/api/v1/auth/patient/login", usecase.PatientHandler(app))
	router.HandleFunc("/api/v1/auth/patient/status", usecase.PatientAuthStatusHandler(app))
	router.HandleFunc("/api/v1/auth/patient/signup", usecase.PatientSignupHandler(app))
	router.HandleFunc("/api/v1/auth/doctor/login", usecase.DoctorHandler(app))
	router.HandleFunc("/api/v1/auth/doctor/signup", usecase.DoctorSignupHandler(app))
	router.HandleFunc("/api/v1/auth/reset-password", usecase.ResetPasswordHandler(app))
	router.HandleFunc("/internal/password-reset/request", usecase.ForgotPasswordHandler(app))
	router.HandleFunc("/api/v1/doctor/list", usecase.DoctorPatientListHandler(app))
	router.HandleFunc("/api/v1/doctors/list", usecase.DoctorsListHandler(app))
	router.HandleFunc("/api/v1/test/grade", usecase.GradeTestHandler(app))
	router.HandleFunc("/api/v1/test/update", usecase.UpdatePatientScoreHandler(app))
	router.HandleFunc("/error", errorHandler)

	return router
}

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	app := &infas.DB_connection{}
	errOpenDb := app.OpenDB()
	if errOpenDb != nil {
		log.Fatalf("can not open a db %v", errOpenDb)
	}

	errInitDb := app.InitDB()
	if errInitDb != nil {
		log.Fatalf("can not open a db %v", errOpenDb)
	}
	server := &http.Server{
		Addr:    ":8080",
		Handler: newRouter(app),
	}

	go func() {
		log.Printf("Server listening on %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	sig := <-stop
	log.Printf("Received signal (%s), shutting down server...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Server shutdown successfully")
}
