package session

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type MockSessionStore struct {
	GetSessionFunc func(sessionID string) (*SessionData, error)
}

func (m *MockSessionStore) GetSession(sessionID string) (*SessionData, error) {
	return m.GetSessionFunc(sessionID)
}

// Implement remaining functions to satisfy interface
func (m *MockSessionStore) CreateSession(userID string, duration time.Duration) (*SessionData, error) {
	return nil, nil
}
func (m *MockSessionStore) DeleteSession(sessionID string) error { return nil }
func (m *MockSessionStore) CleanupExpiredSessions() error        { return nil }

func TestSession_ValidateSession(t *testing.T) {
	mockSession := &Session{
		Store: &MockSessionStore{
			GetSessionFunc: func(sessionID string) (*SessionData, error) {
				if sessionID == "valid-session" {
					return &SessionData{
						ExpiresAt: time.Now().Add(10 * time.Minute),
					}, nil
				}
				if sessionID == "expired-session" {
					return &SessionData{
						ExpiresAt: time.Now().Add(-10 * time.Minute),
					}, nil
				}
				return nil, errors.New("invalid session")
			},
		},
	}

	tests := []struct {
		name        string
		cookieValue string
		expectCode  int
	}{
		{
			name:        "valid session",
			cookieValue: "valid-session",
			expectCode:  http.StatusOK,
		},
		{
			name:        "missing cookie",
			cookieValue: "",
			expectCode:  http.StatusUnauthorized,
		},
		{
			name:        "invalid session",
			cookieValue: "invalid-session",
			expectCode:  http.StatusUnauthorized,
		},
		{
			name:        "expired session",
			cookieValue: "expired-session",
			expectCode:  http.StatusUnauthorized,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if test.cookieValue != "" {
				req.AddCookie(&http.Cookie{
					Name:  CookieName,
					Value: test.cookieValue,
				})
			}

			rr := httptest.NewRecorder()
			handler := mockSession.ValidateSession(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)
			if rr.Code != test.expectCode {
				t.Errorf("expected code %d, got %d", test.expectCode, rr.Code)
			}
		})
	}
}

func TestSession_ValidateAndFetchSession(t *testing.T) {
	mockSession := &Session{
		Store: &MockSessionStore{
			GetSessionFunc: func(sessionID string) (*SessionData, error) {
				if sessionID == "valid-session" {
					return &SessionData{
						ExpiresAt: time.Now().Add(10 * time.Minute),
					}, nil
				}
				if sessionID == "expired-session" {
					return &SessionData{
						ExpiresAt: time.Now().Add(-10 * time.Minute),
					}, nil
				}
				return nil, errors.New("invalid session")
			},
		},
	}

	tests := []struct {
		name        string
		cookieValue string
		expectedErr bool
	}{
		{
			name:        "valid session",
			cookieValue: "valid-session",
			expectedErr: false,
		},
		{
			name:        "missing cookie",
			cookieValue: "",
			expectedErr: true,
		},
		{
			name:        "invalid session",
			cookieValue: "invalid-session",
			expectedErr: true,
		},
		{
			name:        "expired session",
			cookieValue: "expired-session",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.cookieValue != "" {
				req.AddCookie(&http.Cookie{
					Name:  CookieName,
					Value: tt.cookieValue,
				})
			}

			_, err := mockSession.validateAndFetchSession(req)
			if (err != nil) != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err != nil)
			}
		})
	}
}

func TestSetSessionCookie(t *testing.T) {
	tests := []struct {
		name       string
		sessionID  string
		expectName string
		expectPath string
	}{
		{
			name:       "set valid session cookie",
			sessionID:  "session-id-123",
			expectName: CookieName,
			expectPath: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			SetSessionCookie(tt.sessionID, rr)

			cookies := rr.Result().Cookies()
			if len(cookies) != 1 {
				t.Errorf("expected 1 cookie, got %d", len(cookies))
			}

			cookie := cookies[0]
			if cookie.Name != tt.expectName {
				t.Errorf("expected cookie name %s, got %s", tt.expectName, cookie.Name)
			}
			if cookie.Value != tt.sessionID {
				t.Errorf("expected cookie value %s, got %s", tt.sessionID, cookie.Value)
			}
			if !strings.Contains(cookie.Path, tt.expectPath) {
				t.Errorf("expected cookie path %s, got %s", tt.expectPath, cookie.Path)
			}
		})
	}
}

func TestGenerateSessionID(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "generate non-empty session ID"},
		{name: "generate session ID with correct length"},
		{name: "ensure unique session IDs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "generate non-empty session ID":
				sessionID := generateSessionID()
				if sessionID == "" {
					t.Error("expected non-empty session ID, got empty string")
				}

			case "generate session ID with correct length":
				sessionID := generateSessionID()
				if len(sessionID) != 32 {
					t.Errorf("expected session ID length of 32, got %d", len(sessionID))
				}

			case "ensure unique session IDs":
				sessionID1 := generateSessionID()
				sessionID2 := generateSessionID()
				if sessionID1 == sessionID2 {
					t.Error("expected unique session IDs, got identical IDs")
				}
			}
		})
	}
}
