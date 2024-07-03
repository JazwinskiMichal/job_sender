package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"job_sender/interfaces"
	"job_sender/types"
	constants "job_sender/utils/constants"
)

type AuthService struct {
	firebaseWebApiKey     string
	firebaseService       *FirebaseService
	sessionManagerService *SessionManagerService
}

// Ensure firestoreDB conforms to the HashtagDatabase interface.
var _ interfaces.IAuthService = &AuthService{}

// NewAuthService creates a new AuthService backed by Cloud Firestore.
func NewAuthService(firebaseService *FirebaseService, webApiKey string, sessionManagerService *SessionManagerService) *AuthService {
	return &AuthService{
		firebaseWebApiKey:     webApiKey,
		firebaseService:       firebaseService,
		sessionManagerService: sessionManagerService,
	}
}

// Register registers a new user.
func (s *AuthService) Register(email string, password string) (*types.LoginResponseBody, error) {
	requestBody := &types.LoginRequestBody{
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	}
	jsonRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("https://identitytoolkit.googleapis.com/v1/accounts:signUp?key="+s.firebaseWebApiKey, "application/json", bytes.NewBuffer(jsonRequestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var responseBody types.LoginResponseBody
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return nil, err
	}

	return &responseBody, nil
}

// Login logs in a user.
func (s *AuthService) Login(email string, password string) (*types.LoginResponseBody, error) {
	requestBody := &types.LoginRequestBody{
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	}
	jsonRequestBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key="+s.firebaseWebApiKey, "application/json", bytes.NewBuffer(jsonRequestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var responseBody types.LoginResponseBody
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		return nil, err
	}

	return &responseBody, nil
}

// CheckUser returns the user info
func (s *AuthService) CheckUser(r *http.Request) (*types.LoggedUserInfo, error) {
	email, err := s.sessionManagerService.GetElement(r, constants.UserSessionName, constants.UserSessionEmailField)
	if err != nil {
		return nil, err
	}

	// If the email is nil or empty, return an empty LoggedUserInfo
	if email == nil || email == "" {
		return &types.LoggedUserInfo{
			Email:      "",
			IsLoggedIn: false,
			IsVerified: false,
		}, nil
	}

	// Convert the interface to a string
	emailStr, ok := email.(string)
	if !ok {
		return nil, fmt.Errorf("could not convert interface to string")
	}

	isLoggedIn, err := s.checkIsLoggedIn(r)
	if err != nil {
		return nil, err
	}

	isVerified, err := s.firebaseService.CheckIsUserVerified(emailStr)
	if err != nil {
		return nil, err
	}

	return &types.LoggedUserInfo{
		Email:      emailStr,
		IsLoggedIn: isLoggedIn,
		IsVerified: isVerified,
	}, nil
}

// CheckIsLoggedIn checks if a user is logged in.
func (s *AuthService) checkIsLoggedIn(r *http.Request) (bool, error) {
	idToken, err := s.sessionManagerService.GetElement(r, constants.UserSessionName, constants.UserSessionTokenField)
	if err != nil {
		return false, err
	}

	if idToken == nil {
		return false, nil
	}

	// Convert the interface to a string
	idTokenStr, ok := idToken.(string)
	if !ok {
		return false, fmt.Errorf("could not convert interface to string")
	}

	// Get the custom claims
	claims, err := s.firebaseService.GetCustomClaims(idTokenStr)
	if err != nil {
		// Handle expired token error specifically
		if err.Error() == "ID token has expired" {
			return false, nil
		}
		return false, err
	}

	if claims[constants.UserSessionEmailField] == nil {
		return false, nil
	}

	return true, nil
}
