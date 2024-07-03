package interfaces

import (
	"context"

	"firebase.google.com/go/auth"
)

type IFirebaseService interface {
	// CheckIfUserExists checks if the user exists.
	CheckIfUserExists(email string) (bool, error)

	// CheckIsUserVerified checks if the user is verified.
	CheckIsUserVerified(email string) (bool, error)

	// GetCustomClaims retrieves custom claims from Firebase.
	GetCustomClaims(idToken string) (map[string]interface{}, error)

	// Auth is a method that returns the Firebase Auth client.
	Auth(ctx context.Context) (*auth.Client, error)
}
