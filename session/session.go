package session

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"time"
)

const (
	CookieName = "session_id"
)

const (
	unauthorizedMessage    = "Unauthorized access"
	requestCanceledMessage = "Request canceled"
)

type Session struct {
	Store SessionStore
}

// httpError encapsulates an HTTP error response.
type httpError struct {
	message string
	code    int
}

// Error satisfies the error interface for httpError.
func (e httpError) Error() string {
	return e.message
}

// SetSessionCookie sets the cookie to the http response.
func SetSessionCookie(sessionID string, w http.ResponseWriter) {
	// Create a new cookie with the session ID
	cookie := &http.Cookie{
		Name:     CookieName,                         // Cookie name
		Value:    sessionID,                          // Cookie value (session ID)
		Path:     "/",                                // Path for which this cookie is valid
		HttpOnly: true,                               // For security, to make it inaccessible to JS
		Secure:   false,                              // Send only over HTTPS
		Expires:  time.Now().Add(7 * 24 * time.Hour), // Cookie expiration (7 days in this case)
	}

	// Add the cookie to the HTTP response
	http.SetCookie(w, cookie)
}

func (s *Session) ValidateSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionData, err := s.validateAndFetchSession(r)
		if err != nil {
			handleHTTPError(w, err)
			return
		}

		select {
		case <-r.Context().Done():
			http.Error(w, requestCanceledMessage, http.StatusRequestTimeout)
			return
		default:
			ctx := WithSession(r.Context(), sessionData)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}

// validateAndFetchSession validates the session and retrieves session data.
func (s *Session) validateAndFetchSession(r *http.Request) (*SessionData, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil || cookie.Value == "" {
		return nil, httpError{message: unauthorizedMessage, code: http.StatusUnauthorized}
	}

	cookieValue := cookie.Value
	sessionData, err := s.Store.GetSession(cookieValue)
	if err != nil || sessionData.ExpiresAt.Before(time.Now()) {
		return nil, httpError{message: unauthorizedMessage, code: http.StatusUnauthorized}
	}

	return sessionData, nil
}

// handleHTTPError handles HTTP errors by sending the appropriate response.
func handleHTTPError(w http.ResponseWriter, err error) {
	var httpErr httpError
	if errors.As(err, &httpErr) {
		http.Error(w, httpErr.message, httpErr.code)
	}
}

// WithSession attaches a session to a context
func WithSession(ctx context.Context, session *SessionData) context.Context {
	return context.WithValue(ctx, CookieName, session)
}

// GetSessionFromContext retrieves a session from a context
func GetSessionFromContext(ctx context.Context) (*SessionData, bool) {
	session, ok := ctx.Value(CookieName).(*SessionData)
	return session, ok
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
