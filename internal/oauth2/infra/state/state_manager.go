package state

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

type Manager struct {
	states map[string]*Data
	mu     sync.RWMutex
	ttl    time.Duration
}

type Data struct {
	CreatedAt time.Time
	ExpiresAt time.Time
}

func NewStateManager(ttl time.Duration) *Manager {
	sm := &Manager{
		states: make(map[string]*Data),
		ttl:    ttl,
	}

	go sm.clean()

	return sm
}

func (sm *Manager) GenerateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	state := base64.URLEncoding.EncodeToString(b)

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.states[state] = &Data{
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(sm.ttl),
	}

	return state, nil
}

func (sm *Manager) ValidateState(state string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	data, exists := sm.states[state]
	if !exists {
		return exception.ErrStateNotFound
	}

	if time.Now().After(data.ExpiresAt) {
		delete(sm.states, state)
		return exception.ErrStateExpired
	}

	delete(sm.states, state)

	return nil
}

func (sm *Manager) clean() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.mu.Lock()
		now := time.Now()
		for state, data := range sm.states {
			if now.After(data.ExpiresAt) {
				delete(sm.states, state)
			}
		}
		sm.mu.Unlock()
	}
}
