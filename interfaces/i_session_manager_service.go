package interfaces

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

type ISessionManagerService interface {
	// CreateSession creates a new session and stores key-value pairs.
	CreateSession(w http.ResponseWriter, r *http.Request, sessionID string, tokenExpirationTimestamp time.Time, data map[string]interface{}) (*sessions.Session, error)

	// DeleteSession deletes a session.
	DeleteSession(w http.ResponseWriter, r *http.Request, sessionID string) error

	// CheckSession checks if a session exists.
	CheckSession(r *http.Request, sessionID string) bool

	// GetElement retrieves an element from a session.
	GetElement(r *http.Request, sessionID string, key string) (interface{}, error)

	// SetElement sets an element in a session.
	SetElement(w http.ResponseWriter, r *http.Request, sessionID string, key string, value interface{}) error
}
