package core

import (
	"context"
	"fmt"
	"strings"

	"job_sender/interfaces"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

type FirebaseService struct {
	app *firebase.App
}

// Ensure FirebaseService implements IFirebaseService.
var _ interfaces.IFirebaseService = &FirebaseService{}

func NewFirebaseService(secretServiceAccountKey []byte) (*FirebaseService, error) {
	// Create credentials from the service account key
	cred := option.WithCredentialsJSON(secretServiceAccountKey)

	// Initialize Firebase app
	app, err := firebase.NewApp(context.Background(), nil, cred)
	if err != nil {
		return nil, err
	}

	return &FirebaseService{app: app}, nil
}

// CheckIfUserExists checks if the user exists.
func (s *FirebaseService) CheckIfUserExists(email string) (bool, error) {
	ctx := context.Background()

	client, err := s.app.Auth(ctx)
	if err != nil {
		return false, err
	}

	user, err := client.GetUserByEmail(ctx, email)
	if err != nil {
		if strings.Contains(err.Error(), "cannot find user from email") {
			return false, nil
		}
		return false, err
	}

	if user == nil {
		return false, nil
	}

	if user.Email == email {
		return true, nil
	}

	return false, nil
}

// CheckIsUserVerified checks if the user is verified.
func (s *FirebaseService) CheckIsUserVerified(email string) (bool, error) {
	ctx := context.Background()

	client, err := s.app.Auth(ctx)
	if err != nil {
		return false, err
	}

	user, err := client.GetUserByEmail(ctx, email)
	if err != nil {
		return false, err
	}

	return user.EmailVerified, nil
}

// GetCustomClaims retrieves custom claims from Firebase.
func (s *FirebaseService) GetCustomClaims(idToken string) (map[string]interface{}, error) {
	ctx := context.Background()

	client, err := s.app.Auth(ctx)
	if err != nil {
		return nil, err
	}

	token, err := client.VerifyIDToken(ctx, idToken)
	if err != nil {
		// Check if the error is due to an expired token
		if strings.Contains(err.Error(), "has expired at") {
			return nil, fmt.Errorf("ID token has expired")
		}
		return nil, err
	}

	return token.Claims, nil
}

// Auth is a method that returns the Firebase Auth client.
func (s *FirebaseService) Auth(ctx context.Context) (*auth.Client, error) {
	return s.app.Auth(ctx)
}
