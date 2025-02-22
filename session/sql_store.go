package session

import (
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite" // SQLite driver
)

// DBSessionStore is an SQL-based implementation of the SessionStore interface
type DBSessionStore struct {
	db *sql.DB
}

func NewDBSessionStore(dsn string, driver string) (*DBSessionStore, error) {
	db, err := sql.Open(driver, dsn) // Pass driver (e.g., "sqlite3", "postgres", etc.)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			expires_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return nil, err
	}

	return &DBSessionStore{db: db}, nil
}

// CreateSession creates a new session and stores it in the database
func (s *DBSessionStore) CreateSession(userID string, duration time.Duration) (*SessionData, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	session := &SessionData{
		ID:        generateSessionID(),
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(duration),
	}

	_, err := s.db.Exec(`
		INSERT INTO sessions (id, user_id, created_at, expires_at)
		VALUES (?, ?, ?, ?)
	`, session.ID, session.UserID, session.CreatedAt, session.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetSession retrieves a session by its ID
func (s *DBSessionStore) GetSession(sessionID string) (*SessionData, error) {
	row := s.db.QueryRow(`
		SELECT id, user_id, created_at, expires_at
		FROM sessions
		WHERE id = ?
	`, sessionID)

	var session SessionData
	err := row.Scan(&session.ID, &session.UserID, &session.CreatedAt, &session.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("session not found")
	} else if err != nil {
		return nil, err
	}

	if session.ExpiresAt.Before(time.Now()) {
		_ = s.DeleteSession(sessionID)
		return nil, errors.New("session expired")
	}

	return &session, nil
}

// DeleteSession deletes a session by its ID
func (s *DBSessionStore) DeleteSession(sessionID string) error {
	_, err := s.db.Exec(`
		DELETE FROM sessions
		WHERE id = ?
	`, sessionID)
	return err
}

// CleanupExpiredSessions removes all expired sessions from the database
func (s *DBSessionStore) CleanupExpiredSessions() error {
	_, err := s.db.Exec(`
		DELETE FROM sessions
		WHERE expires_at < ?
	`, time.Now())
	return err
}
