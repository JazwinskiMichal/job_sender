package types

type EnvVariables struct {
	Port string

	ProjectID         string
	ProjectLocationID string
	ProjectNumber     string

	SecretNameServiceAccountKey       string
	SecretNameFirestoreWebApiKey      string
	SecretNameEmailServiceEmail       string
	SecretNameEmailServiceAppPassword string
	SecretNameSessionCookieStore      string

	EmailAggregatorQueueName string

	TimesheetsBucketName string
}
