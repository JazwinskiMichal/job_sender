package core

import (
	"job_sender/interfaces"
	"job_sender/types"
	"log"
	"os"
)

type EnvVariablesService struct {
	portKey string

	projectIDKey       string
	projectLocationKey string
	projectNumberKey   string

	secretNameServiceAccountKey          string
	secretNameFirestoreWebApiKey         string
	secretNameEmailServiceEmailKey       string
	secretNameEmailServiceAppPasswordKey string
	secretNameSessionCookieStoreKey      string

	emailAggregatorQueueNameKey string

	timeSheetsBucketNameKey string
}

// Ensure EnvVariablesService implements IEnvVariablesService.
var _ interfaces.IEnvVariablesService = &EnvVariablesService{}

func NewEnvVariablesService(portKey string, projectIDKey string, projectLocationIDKey string, projectNumberKey string,
	secretNameServiceAccountKey string, secretNameFirestoreWebApiKey string, secretNameEmailServiceEmailKey string, secretNameEmailServiceAppPasswordKey string, secretNameSessionCookieStoreKey string,
	emailAggregatorQueueNameKey string,
	timesheetsBucketNameKey string) *EnvVariablesService {

	return &EnvVariablesService{
		portKey: portKey,

		projectIDKey:       projectIDKey,
		projectLocationKey: projectLocationIDKey,
		projectNumberKey:   projectNumberKey,

		secretNameServiceAccountKey:          secretNameServiceAccountKey,
		secretNameFirestoreWebApiKey:         secretNameFirestoreWebApiKey,
		secretNameEmailServiceEmailKey:       secretNameEmailServiceEmailKey,
		secretNameEmailServiceAppPasswordKey: secretNameEmailServiceAppPasswordKey,
		secretNameSessionCookieStoreKey:      secretNameSessionCookieStoreKey,

		emailAggregatorQueueNameKey: emailAggregatorQueueNameKey,

		timeSheetsBucketNameKey: timesheetsBucketNameKey,
	}
}

// GetEnvVariables gets the environment variables from the cloudbuild.yaml file.
func (e *EnvVariablesService) GetEnvVariables() *types.EnvVariables {
	port := os.Getenv(e.portKey)
	if port == "" {
		port = "8080"
	}

	projectID := os.Getenv(e.projectIDKey)
	if projectID == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT_ID must be set")
	}

	projectLocationID := os.Getenv(e.projectLocationKey)
	if projectLocationID == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT_LOCATION_ID must be set")
	}

	projectNumber := os.Getenv(e.projectNumberKey)
	if projectNumber == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT_NUMBER must be set")
	}

	secretNameServiceAccountKey := os.Getenv(e.secretNameServiceAccountKey)
	if secretNameServiceAccountKey == "" {
		log.Fatal("SECRET_NAME_SERVICE_ACCOUNT_KEY must be set")
	}

	secretNameFirestoreWebApiKey := os.Getenv(e.secretNameFirestoreWebApiKey)
	if secretNameFirestoreWebApiKey == "" {
		log.Fatal("SECRET_NAME_FIREBASE_WEB_API_KEY must be set")
	}

	secretNameEmailServiceEmail := os.Getenv(e.secretNameEmailServiceEmailKey)
	if secretNameEmailServiceEmail == "" {
		log.Fatal("SECRET_NAME_EMAIL_SERVICE_EMAIL must be set")
	}

	secretNameEmailServiceAppPassword := os.Getenv(e.secretNameEmailServiceAppPasswordKey)
	if secretNameEmailServiceAppPassword == "" {
		log.Fatal("SECRET_NAME_EMAIL_SERVICE_APP_PASSWORD must be set")
	}

	secretNameSessionCookieStore := os.Getenv(e.secretNameSessionCookieStoreKey)
	if secretNameSessionCookieStore == "" {
		log.Fatal("SECRET_NAME_SESSION_COOKIE_STORE must be set")
	}

	emailAggregatorQueueName := os.Getenv(e.emailAggregatorQueueNameKey)
	if emailAggregatorQueueName == "" {
		log.Fatal("EMAIL_AGGREGATOR_QUEUE_NAME must be set")
	}

	timesheetsBucketName := os.Getenv(e.timeSheetsBucketNameKey)
	if timesheetsBucketName == "" {
		log.Fatal("TIMESHEETS_BUCKET_NAME must be set")
	}

	return &types.EnvVariables{
		Port: port,

		ProjectID:         projectID,
		ProjectLocationID: projectLocationID,
		ProjectNumber:     projectNumber,

		SecretNameServiceAccountKey:       secretNameServiceAccountKey,
		SecretNameFirestoreWebApiKey:      secretNameFirestoreWebApiKey,
		SecretNameEmailServiceEmail:       secretNameEmailServiceEmail,
		SecretNameEmailServiceAppPassword: secretNameEmailServiceAppPassword,
		SecretNameSessionCookieStore:      secretNameSessionCookieStore,

		EmailAggregatorQueueName: emailAggregatorQueueName,

		TimesheetsBucketName: timesheetsBucketName,
	}
}
