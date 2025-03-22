package session

import (
	"testing"
	"time"
)

func TestInMemorySessionStore_CreateSession(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		duration  time.Duration
		wantError bool
	}{
		{"valid_session_short_duration", "user1", time.Minute, false},
		{"valid_session_long_duration", "user2", time.Hour, false},
		{"empty_user_id", "", time.Minute, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewInMemorySessionStore()
			session, err := store.CreateSession(tt.userID, tt.duration)

			if (err != nil) != tt.wantError {
				t.Fatalf("CreateSession() error = %v, wantError %v", err, tt.wantError)
			}

			if session == nil {
				t.Fatal("Expected session to be created but got nil")
			}

			if session.UserID != tt.userID {
				t.Errorf("Expected userID = %v, got = %v", tt.userID, session.UserID)
			}

			if session.ExpiresAt.Sub(session.CreatedAt).Round(time.Second) != tt.duration {
				t.Errorf("Expected session duration = %v, got = %v", tt.duration, session.ExpiresAt.Sub(session.CreatedAt))
			}
		})
	}
}

func TestInMemorySessionStore_GetSession(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		setup     func(store *InMemorySessionStore) string
		wantError bool
	}{
		{
			"valid_session",
			"",
			func(store *InMemorySessionStore) string {
				session, _ := store.CreateSession("user1", time.Minute)
				return session.ID
			},
			false,
		},
		{
			"expired_session",
			"",
			func(store *InMemorySessionStore) string {
				session, _ := store.CreateSession("user1", -time.Minute)
				return session.ID
			},
			true,
		},
		{
			"nonexistent_session",
			"invalidID",
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewInMemorySessionStore()
			var id string
			if tt.setup != nil {
				id = tt.setup(store)
			} else {
				id = tt.sessionID
			}

			session, err := store.GetSession(id)

			if (err != nil) != tt.wantError {
				t.Fatalf("GetSession() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && session == nil {
				t.Fatal("Expected valid session but got nil")
			}

			if tt.wantError && session != nil {
				t.Fatal("Expected no session but got a valid session")
			}
		})
	}
}

func TestInMemorySessionStore_DeleteSession(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		setup     func(store *InMemorySessionStore) string
	}{
		{
			"delete_existing_session",
			"",
			func(store *InMemorySessionStore) string {
				session, _ := store.CreateSession("user1", time.Minute)
				return session.ID
			},
		},
		{
			"delete_nonexistent_session",
			"invalidID",
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewInMemorySessionStore()
			var id string
			if tt.setup != nil {
				id = tt.setup(store)
			} else {
				id = tt.sessionID
			}

			err := store.DeleteSession(id)
			if err != nil {
				t.Fatalf("DeleteSession() error = %v", err)
			}

			_, err = store.GetSession(id)
			if err == nil && id != "invalidID" {
				t.Fatal("Expected session to be deleted but it still exists")
			}
		})
	}
}

func TestInMemorySessionStore_CleanupExpiredSessions(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(store *InMemorySessionStore)
		expectCount int
	}{
		{
			"no_expired_sessions",
			func(store *InMemorySessionStore) {
				_, _ = store.CreateSession("user1", time.Minute)
				_, _ = store.CreateSession("user2", time.Minute)
			},
			2,
		},
		{
			"some_expired_sessions",
			func(store *InMemorySessionStore) {
				_, _ = store.CreateSession("user1", -time.Minute)
				_, _ = store.CreateSession("user2", time.Minute)
			},
			1,
		},
		{
			"all_expired_sessions",
			func(store *InMemorySessionStore) {
				_, _ = store.CreateSession("user1", -time.Minute)
				_, _ = store.CreateSession("user2", -time.Minute)
			},
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewInMemorySessionStore()
			tt.setup(store)

			_ = store.CleanupExpiredSessions()

			store.mutex.RLock()
			defer store.mutex.RUnlock()
			if len(store.sessions) != tt.expectCount {
				t.Errorf("Expected session count = %v, got = %v", tt.expectCount, len(store.sessions))
			}
		})
	}
}
