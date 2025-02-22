package session

import (
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *DBSessionStore {
	t.Helper()

	store, err := NewDBSessionStore(":memory:", "sqlite")
	if err != nil {
		t.Fatalf("failed to create test DB: %v", err)
	}

	return store
}

func TestDBSessionStore_CreateSession(t *testing.T) {
	store := setupTestDB(t)

	tests := []struct {
		name     string
		userID   string
		duration time.Duration
		wantErr  bool
	}{
		{"valid session", "user1", 1 * time.Hour, false},
		{"empty user ID", "", 1 * time.Hour, true},
		{"negative duration", "user2", -1 * time.Hour, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := store.CreateSession(tt.userID, tt.duration)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSession() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && session == nil {
				t.Error("CreateSession() returned nil session")
			}
		})
	}
}

func TestDBSessionStore_GetSession(t *testing.T) {
	store := setupTestDB(t)

	expiredSession, _ := store.CreateSession("user3", -1*time.Second)
	validSession, _ := store.CreateSession("user4", 1*time.Hour)

	tests := []struct {
		name      string
		sessionID string
		wantErr   string
	}{
		{"valid session", validSession.ID, ""},
		{"expired session", expiredSession.ID, "session expired"},
		{"nonexistent session", "nonexistent", "session not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := store.GetSession(tt.sessionID)
			if err != nil && err.Error() != tt.wantErr {
				t.Errorf("GetSession() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == "" && session == nil {
				t.Errorf("GetSession() returned nil for a valid session")
			}
		})
	}
}

func TestDBSessionStore_DeleteSession(t *testing.T) {
	store := setupTestDB(t)

	session, _ := store.CreateSession("user5", 1*time.Hour)

	tests := []struct {
		name      string
		sessionID string
		wantErr   bool
	}{
		{"existing session", session.ID, false},
		{"nonexistent session", "nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.DeleteSession(tt.sessionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteSession() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr == false {
				_, err := store.GetSession(tt.sessionID)
				if err == nil {
					t.Errorf("DeleteSession() did not remove session")
				}
			}
		})
	}
}

func TestDBSessionStore_CleanupExpiredSessions(t *testing.T) {
	store := setupTestDB(t)

	sessionExpired, _ := store.CreateSession("user6", -1*time.Second)
	sessionExists, _ := store.CreateSession("user7", 1*time.Hour)

	err := store.CleanupExpiredSessions()
	if err != nil {
		t.Fatalf("CleanupExpiredSessions() error = %v", err)
	}

	tests := []struct {
		name      string
		sessionID string
		wantErr   string
	}{
		{"expired session", sessionExpired.ID, "session not found"},
		{"valid session", sessionExists.ID, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := store.GetSession(tt.sessionID)
			if (err != nil && err.Error() != tt.wantErr) || (err == nil && tt.wantErr != "") {
				t.Errorf("CleanupExpiredSessions() state incorrect for session ID %v", tt.sessionID)
			}
		})
	}
}
