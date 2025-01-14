package core

import (
	"net/http"
	"time"

	"job_sender/interfaces"

	"github.com/gorilla/sessions"
)

type SessionManagerService struct {
	store *sessions.CookieStore
}

// Ensure SessionManagerService implements ISessionManagerService.
var _ interfaces.ISessionManagerService = &SessionManagerService{}

// NewSessionManagerService creates a new SessionManagerService.
func NewSessionManagerService(secretCookieStoreKey []byte) *SessionManagerService {
	store := sessions.NewCookieStore(secretCookieStoreKey)

	return &SessionManagerService{
		store: store,
	}
}

// CreateSession creates a new session and stores key-value pairs.
func (s *SessionManagerService) CreateSession(w http.ResponseWriter, r *http.Request, sessionID string, tokenExpirationTimestamp time.Time, data map[string]interface{}) (*sessions.Session, error) {
	// Initialize a new session for the user
	session, err := s.store.Get(r, sessionID)
	if err != nil {
		return nil, err
	}

	// Store the data in the session
	for key, value := range data {
		session.Values[key] = value
	}

	// Calculate MaxAge based on tokenExpirationTimestamp
	currentTime := time.Now()
	maxAge := int(tokenExpirationTimestamp.Sub(currentTime).Seconds())

	// Set the session options
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   maxAge, // TODO: set low maxAge and test, what would happen when session ends
		HttpOnly: true,
		Secure:   true,                 // Only send cookie over HTTPS
		SameSite: http.SameSiteLaxMode, // Adjust according to your requirements
	}

	// Save the session
	err = session.Save(r, w)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// DeleteSession deletes a session.
func (s *SessionManagerService) DeleteSession(w http.ResponseWriter, r *http.Request, sessionID string) error {
	session, err := s.store.Get(r, sessionID)
	if err != nil {
		return err
	}

	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		return err
	}

	return nil
}

// CheckSession checks if a session exists.
func (s *SessionManagerService) CheckSession(r *http.Request, sessionID string) bool {
	session, err := s.store.Get(r, sessionID)
	if err != nil {
		return false
	}

	return session.IsNew
}

// GetElement retrieves an element from a session.
func (s *SessionManagerService) GetElement(r *http.Request, sessionID string, key string) (interface{}, error) {
	session, err := s.store.Get(r, sessionID)
	if err != nil {
		return nil, err
	}

	return session.Values[key], nil
}

// SetElement sets an element in a session.
func (s *SessionManagerService) SetElement(w http.ResponseWriter, r *http.Request, sessionID string, key string, value interface{}) error {
	session, err := s.store.Get(r, sessionID)
	if err != nil {
		return err
	}

	session.Values[key] = value
	err = session.Save(r, w)
	if err != nil {
		return err
	}

	return nil
}
