package interfaces

import (
	"net/http"
	"time"
)

type ISessionManagerService interface {
	// CreateSession creates a new session and stores key-value pairs.
	CreateSession(w http.ResponseWriter, r *http.Request, sessionID string, tokenExpirationTimestamp time.Time, data map[string]interface{}) error

	// DeleteSession deletes a session.
	DeleteSession(w http.ResponseWriter, r *http.Request, sessionID string) error

	// GetElement retrieves an element from a session.
	GetElement(r *http.Request, sessionID string, key string) (interface{}, error)
}
