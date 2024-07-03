package core

import (
	"context"
	"fmt"

	"job_sender/interfaces"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// SecretManagerService struct
type SecretManagerService struct {
	client *secretmanager.Client
}

// Ensure SecretManagerClient conforms to the ISecretManagerClient interface.
var _ interfaces.ISecretManagerService = &SecretManagerService{}

// NewSecretManagerService creates a new SecretManagerClient
func NewSecretManagerService() (*SecretManagerService, error) {
	client, err := secretmanager.NewClient(context.Background())
	if err != nil {
		return nil, err
	}
	return &SecretManagerService{client: client}, nil
}

// GetSecret retrieves a secret from Google Cloud Secret Manager
func (s *SecretManagerService) GetSecret(projectNumber string, kmsSecretName string) ([]byte, error) {
	// Build the request
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectNumber, kmsSecretName),
	}

	// Call the Secret Manager API to get the secret
	result, err := s.client.AccessSecretVersion(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return result.Payload.Data, nil
}
