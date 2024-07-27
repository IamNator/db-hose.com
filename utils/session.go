package utils

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	sessionDuration      = 2 * time.Hour
	sessionInactivityMax = 30 * time.Minute
)

type Session struct {
	Email      string
	LastAccess time.Time
}

type SessionManager struct {
	sessions map[string]*Session
	mu       sync.Mutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

// InitializeSessionCleaner initializes a background goroutine to clean up expired sessions
func (sm *SessionManager) InitializeSessionCleaner() {
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			sm.Cleanup()
		}
	}()
}

// CreateSession creates a new session for the given username and returns the session token
func (sm *SessionManager) CreateSession(email string) (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	token, err := GenerateJWT(email, sessionDuration)
	if err != nil {
		return "", err
	}

	sm.sessions[token] = &Session{
		Email:      email,
		LastAccess: time.Now(),
	}

	return token, nil
}

// GetSession retrieves the session by token and refreshes the last access time
func (sm *SessionManager) GetSession(token string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[token]
	if !exists {
		return nil, errors.New("session not found")
	}

	if time.Since(session.LastAccess) > sessionInactivityMax {
		delete(sm.sessions, token)
		return nil, errors.New("session expired due to inactivity")
	}

	if time.Since(session.LastAccess) > sessionDuration {
		delete(sm.sessions, token)
		return nil, errors.New("session expired")
	}

	session.LastAccess = time.Now()
	return session, nil
}

// DeleteSession deletes the session by token
func (sm *SessionManager) DeleteSession(token string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, token)
}

// Cleanup removes expired sessions from the session map
func (sm *SessionManager) Cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for token, session := range sm.sessions {
		if time.Since(session.LastAccess) > sessionInactivityMax || time.Since(session.LastAccess) > sessionDuration {
			delete(sm.sessions, token)
		}
	}
}

func (sm *SessionManager) ValidateSession(token string) error {
	_, err := sm.GetSession(token)
	return err
}

func (sm *SessionManager) Middleware(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	if err := sm.ValidateSession(token); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	c.Set("email", sm.sessions[token].Email)
	c.Next()
}
