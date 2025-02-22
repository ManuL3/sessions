package session

import (
	"math/rand"
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

// generateSessionID generates a random session ID
func generateSessionID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
