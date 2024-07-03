package interfaces

import (
	"job_sender/types"
	"net/http"
)

// IAuthService provides functions to authenticate a user.
type IAuthService interface {
	// Register registers a new user.
	Register(email string, password string) (*types.LoginResponseBody, error)

	// Login logs in a user.
	Login(email string, password string) (*types.LoginResponseBody, error)

	// CheckUser returns the user info
	CheckUser(r *http.Request) (*types.LoggedUserInfo, error)
}
