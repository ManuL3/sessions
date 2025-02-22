package session

import (
	"net/http"
	"time"
)

const (
	CookieName = "session_id"
)

type Session struct {
	Store SessionStore
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
