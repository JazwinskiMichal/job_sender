package types

type EnvVariables struct {
	Port          string
	ProjectID     string
	ProjectNumber string

	SecretNameServiceAccountKey       string
	SecretNameFirestoreWebApiKey      string
	SecretNameEmailServiceEmail       string
	SecretNameEmailServiceAppPassword string
	SecretNameSessionCookieStore      string
}
