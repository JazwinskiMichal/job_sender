package interfaces

type ISecretManagerService interface {
	// GetSecret retrieves a secret from Google Cloud Secret Manager by project number and secret name.
	GetSecret(projectNumber string, kmsSecretName string) ([]byte, error)
}
