package session

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// httpError encapsulates an HTTP error response.
type httpError struct {
	message string
	code    int
}

// Error satisfies the error interface for httpError.
func (e httpError) Error() string {
	return e.message
}

const (
	unauthorizedMessage    = "Unauthorized access"
	requestCanceledMessage = "Request canceled"
)

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
