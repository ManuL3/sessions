package session

import (
	"errors"
	"sync"
	"time"
)

type InMemorySessionStore struct {
	sessions map[string]*SessionData
	mutex    sync.RWMutex
}

func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		sessions: make(map[string]*SessionData),
	}
}

func (s *InMemorySessionStore) CreateSession(userID string, duration time.Duration) (*SessionData, error) {
	session := &SessionData{
		ID:        generateSessionID(),
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(duration),
	}

	s.mutex.Lock()
	s.sessions[session.ID] = session
	s.mutex.Unlock()

	return session, nil
}

func (s *InMemorySessionStore) GetSession(sessionID string) (*SessionData, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists || session.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("session not found or expired")
	}

	return session, nil
}

func (s *InMemorySessionStore) DeleteSession(sessionID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.sessions, sessionID)
	return nil
}

func (s *InMemorySessionStore) CleanupExpiredSessions() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for id, session := range s.sessions {
		if session.ExpiresAt.Before(time.Now()) {
			delete(s.sessions, id)
		}
	}

	return nil
}
