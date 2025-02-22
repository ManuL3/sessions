package main

import (
	"encoding/gob"
	"github.com/ManuL3/sessions/session"
	"log"
	"net/http"
	"time"
)

func main() {
	// Choose a session store (e.g., SQL or in-memory)
	store, err := session.NewDBSessionStore("sessions.db", "sqlite")
	if err != nil {
		log.Fatalf("Failed to initialize session store: %v", err)
	}

	sessionCtrl := session.Session{}
	sessionCtrl.Store = store

	gob.Register(session.SessionData{})

	// Example HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/login", LoginHandler(store))
	mux.Handle("/protected", sessionCtrl.ValidateSession(ProtectedHandler()))

	log.Println("Serving on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func LoginHandler(store session.SessionStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := store.CreateSession("user123", 30*time.Minute)
		if err != nil {
			http.Error(w, "Error creating session", http.StatusInternalServerError)
			return
		}
		session.SetSessionCookie(s.ID, w)
		w.Write([]byte("Session: " + s.ID))
	}
}

func ProtectedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		sessionData := r.Context().Value(session.CookieName)
		log.Println(sessionData)
		if sessionData == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		sessionDataConverted := sessionData.(*session.SessionData)

		if sessionDataConverted.UserID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Write([]byte("Hello, " + sessionDataConverted.UserID))
	}
}
