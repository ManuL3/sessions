## Contributing 

This project is intended as a personal initiative, shared primarily for use in other personal Go projects. While the
code is publicly accessible, contributions, suggestions, or pull requests are not required or expected. 


## Project Overview

This project provides a **session management middleware** for Go web applications. It includes session validation
functionalities and utility methods to manage session data efficiently. The middleware is built to integrate seamlessly
with HTTP handlers in a web server, allowing session validation, storage, and retrieval of session details from HTTP
requests.
The project is designed with modularity and readability in mind, adhering to Go idiomatic principles. The core features
include middleware to validate sessions, helper functions to embed session data in the context, and utility methods to
manage session lifecycles.

### Features

- **Middleware for Session Validation**:
    - Automatically validates incoming requests by checking session headers.
    - Ensures that requests without valid sessions are rejected with appropriate HTTP error codes.
    - Handles expired sessions gracefully.

- **Utility Methods**:
    - Embed session data in the request's context for downstream processing.
    - Retrieve session data from the request's context wherever required.

- **Session Store Interface**:
    - Supports session management through `CreateSession`, `GetSession`, `DeleteSession`, and `CleanupExpiredSessions`
      methods.

### API Documentation

#### Session Middleware

``` go
func (s *SessionStore) ValidateSession(next http.Handler) http.Handler
```

- **Purpose**: Validates session by verifying the presence of an HTTP cookie and fetching the corresponding session data
  from the session store.
  It ensures that the session is valid and not expired.

- **Parameters**:
    - `next http.Handler`: The next handler in the middleware chain.

- **Usage**: Attach this middleware to your HTTP server to enforce session validation.

#### Context Helpers

``` go
func attachSessionToContext(ctx context.Context, session *SessionData) context.Context
```

- **Purpose**: Embeds session data into a given request's context.
- **Parameters**:
    - `ctx`: The existing context.
    - `session *SessionData`: The session data to embed.

- **Returns**: A new context containing the session data.

``` go
func getSessionFromContext(ctx context.Context) (*SessionData, bool)
```

- **Purpose**: Retrieves session data from the context.
- **Parameters**:
    - `ctx`: The context containing session data.

- **Returns**:
    - A `*SessionData` pointer if the session exists in the context.
    - A boolean indicating whether the session data was found.

#### Error Response Helper

``` go
func respondWithError(w http.ResponseWriter, status int, message string)
```

- **Purpose**: Sends a standardized HTTP error response.
- **Parameters**:
    - `w`: The `http.ResponseWriter` to write the error response.
    - `status`: HTTP status code (e.g., `http.StatusUnauthorized`).
    - `message`: Error message to be sent in the response.

#### Context Cancellation Handler

``` go
func handleContextCancel(w http.ResponseWriter, ctx context.Context) bool
```

- **Purpose**: Checks if the request context has been canceled. If canceled, sends an HTTP error response and stops
  processing.
- **Parameters**:
    - `w`: The `http.ResponseWriter` to write the error response.
    - `ctx`: The request's context.

- **Returns**: A boolean indicating whether the context was canceled.

### Getting Started

#### Prerequisites:

- **Go**: Ensure you have Go installed (minimum version >= 1.23).

#### Installation:

1. Clone the repository:

``` bash
   git clone https://github.com/ManuL3/sessions.git
   cd sessions
```

1. Install dependencies (if any modules are declared):

``` bash
   go mod tidy
```

1. Implement the `SessionStore` interface with your desired session storage mechanism (e.g., database, in-memory store).

### Usage

Below is an example of how you can integrate the session middleware into a typical Go HTTP server:

``` go
package main

import (
	"net/http"
	"mrSession/session"
)

// Mock SessionStore Implementation
type InMemorySessionStore struct {}
func (s *InMemorySessionStore) GetSession(sessionID string) (*session.SessionData, error) {
	// Mock implementation for returning session data
	return &session.SessionData{
		ID:        sessionID,
		UserID:    "user123",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}, nil
}

func main() {
	sessionStore := &InMemorySessionStore{}

	mux := http.NewServeMux()
	mux.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
		// Retrieve session from context
		sessionData, ok := session.getSessionFromContext(r.Context())
		if !ok {
			http.Error(w, "No session found in context", http.StatusInternalServerError)
			return
		}

		w.Write([]byte("Hello, " + sessionData.UserID))
	})

	// Apply middleware
	http.ListenAndServe(":8080", sessionStore.ValidateSession(mux))
}
```

### Project Structure

``` plaintext
/session
  ├── session.go              // Core middleware and helper methods
  ├── session_store.go        // Definition of the `SessionStore` interface
  └── README.md               // Documentation of the project
```


### License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
