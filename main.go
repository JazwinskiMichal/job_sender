package main

import (
	"log"
	"net/http"

	"job_sender/core"
	"job_sender/handlers"
	"job_sender/middlewares"
)

func main() {
	// Create new EnvVariablesService
	envVariablesService := core.NewEnvVariablesService("PORT", "GOOGLE_CLOUD_PROJECT_ID", "GOOGLE_CLOUD_PROJECT_LOCATION_ID", "GOOGLE_CLOUD_PROJECT_NUMBER", "SERVICE_ACCOUNT_EMAIL", "SECRET_NAME_SERVICE_ACCOUNT_KEY", "SECRET_NAME_FIREBASE_WEB_API_KEY", "SECRET_NAME_EMAIL_SERVICE_EMAIL", "SECRET_NAME_EMAIL_SERVICE_APP_PASSWORD", "SECRET_NAME_SESSION_COOKIE_STORE", "EMAIL_AGGREGATOR_QUEUE_NAME", "TIMESHEETS_BUCKET_NAME")
	envVariables := envVariablesService.GetEnvVariables()

	// Create a new Secret Manager client
	s, err := core.NewSecretManagerService()
	if err != nil {
		log.Fatalf("Failed to create secret manager client: %v", err)
	}

	// Get the sa-backend service account key from Secret Manager
	secretServiceAccountKey, err := s.GetSecret(envVariables.ProjectNumber, envVariables.SecretNameServiceAccountKey)
	if err != nil {
		log.Fatalf("Failed to get secret: %v", err)
	}

	// Initialize Firebase service
	firebaseService, err := core.NewFirebaseService(secretServiceAccountKey)
	if err != nil {
		log.Fatalf("NewFirebaseService: %v", err)
	}

	// Get the Firestore Web API key from Secret Manager
	firebaseWebApiKey, err := s.GetSecret(envVariables.ProjectNumber, envVariables.SecretNameFirestoreWebApiKey)
	if err != nil {
		log.Fatalf("Failed to get secret: %v", err)
	}

	// Get the email and app password for the email service from Secret Manager
	emailServiceEmail, err := s.GetSecret(envVariables.ProjectNumber, envVariables.SecretNameEmailServiceEmail)
	if err != nil {
		log.Fatalf("Failed to get secret: %v", err)
	}
	emailServiceAppPassword, err := s.GetSecret(envVariables.ProjectNumber, envVariables.SecretNameEmailServiceAppPassword)
	if err != nil {
		log.Fatalf("Failed to get secret: %v", err)
	}

	// Initialize Error Reporter Service
	errorReporterService := core.NewErrorReporterService(envVariables.ProjectID)
	if errorReporterService == nil {
		log.Fatalf("NewErrorReporterService: %v", err)
	}

	// Initialize Email Service
	emailService := core.NewEmailService(string(emailServiceEmail), string(emailServiceAppPassword))

	// Initialize Panic Recover Middleware
	panicRecoverMiddleware := middlewares.NewPanicRecoverMiddleware(errorReporterService)

	// Initialize Template Service
	templateService := core.NewTemplateService()

	// Get the session cookie store secret from Secret Manager
	sessionCookieStore, err := s.GetSecret(envVariables.ProjectNumber, envVariables.SecretNameSessionCookieStore)
	if err != nil {
		log.Fatalf("Failed to get secret: %v", err)
	}

	// Initialize Sesssion Manager Service
	sessionManagerService := core.NewSessionManagerService(sessionCookieStore)

	// Initialize Cloud Tasks service
	cloudTasksService := core.NewCloudTasksService(envVariablesService)

	// Initialize Cloud Scheduler Service
	schedulerService, err := core.NewSchedulerService(envVariables.ServiceAccountEmail, envVariables.ProjectID, envVariables.ProjectLocationID, secretServiceAccountKey)

	// Initialize the Auth service
	authService := core.NewAuthService(firebaseService, string(firebaseWebApiKey), sessionManagerService)
	authMiddleware := middlewares.NewAuthMiddleware(authService, errorReporterService)

	// Create Storage Service
	storageService, err := core.NewStorageService(envVariables.TimesheetsBucketName)
	if err != nil {
		log.Fatalf("NewStorageService: %v", err)
	}

	// Create owners db service
	ownersDB, err := core.NewOwnerDatabaseService(firebaseService)
	if err != nil {
		log.Fatalf("NewOwnerDatabaseService: %v", err)
	}

	// Create new Main handler and router
	mainHandler := handlers.NewMainHandler(authService, ownersDB, errorReporterService)

	// Create the router
	router := mainHandler.CreateRouter()
	router.Use(panicRecoverMiddleware.PanicRecoverMiddleware)

	// Create a subrouter for routes that require authentication
	authRouter := router.PathPrefix("/auth").Subrouter()
	authRouter.Use(authMiddleware.AuthMiddleware)

	// Create register handler
	registerHandler := handlers.NewRegisterHandler(firebaseService, authService, templateService, emailService, errorReporterService)
	registerHandler.RegisterRegisterHandlers(router)

	// Create login handler
	loginHandler := handlers.NewLoginHandler(authService, firebaseService, templateService, sessionManagerService, errorReporterService)
	loginHandler.RegisterLoginHandlers(router)

	// Create Something went wrong handler
	somethingWentWrongHandler := handlers.NewSomethingWentWrongHandler(templateService)
	somethingWentWrongHandler.RegisterSomethingWentWrongHandlers(router)

	// Create the groups db service
	groupsDB, err := core.NewGroupsDatabaseService(firebaseService)
	if err != nil {
		log.Fatalf("NewGroupsDatabaseService: %v", err)
	}

	// Create contractors db service
	contractorsDB, err := core.NewContractorsDatabaseService(firebaseService)
	if err != nil {
		log.Fatalf("NewContractorsDatabaseService: %v", err)
	}

	// Create timesheets db service
	timesheetsDB, err := core.NewTimesheetsDatabaseService(firebaseService)
	if err != nil {
		log.Fatalf("NewTimesheetsDatabaseService: %v", err)
	}

	// Create owners handler
	ownersHandler := handlers.NewOwnersHandler(authService, ownersDB, sessionManagerService, templateService, errorReporterService)
	ownersHandler.RegisterOwnersHandlers(authRouter)

	// Create groups handler
	groupsHandler := handlers.NewGroupsHandler(authService, ownersDB, groupsDB, schedulerService, sessionManagerService, templateService, errorReporterService)
	groupsHandler.RegisterGroupsHandlers(authRouter)

	// Create contractor handler
	contractorsHandler := handlers.NewContractorsHandler(authService, groupsDB, contractorsDB, cloudTasksService, templateService, errorReporterService, envVariables)
	contractorsHandler.RegisterContractorsHandler(authRouter)

	// Create timesheets handler
	timesheetsHandler := handlers.NewTimesheetsHandler(emailService, contractorsDB, timesheetsDB, storageService, errorReporterService)
	timesheetsHandler.RegisterTimesheetsHandlers(router)

	// Start the server
	if err := http.ListenAndServe(":"+envVariables.Port, router); err != nil {
		log.Fatal(err)
	}
}

// TODO: dać możliwość na liście kontraktowców do ręcznego wyzwolenia timesheet request
