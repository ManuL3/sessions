package session

import (
	"time"
)

// SessionData represents the structure of a session
type SessionData struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// SessionStore defines an interface for session storage backends
type SessionStore interface {
	CreateSession(userID string, duration time.Duration) (*SessionData, error)
	GetSession(sessionID string) (*SessionData, error)
	DeleteSession(sessionID string) error
	CleanupExpiredSessions() error
}
