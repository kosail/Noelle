package main

import (
	"sync"
	"time"

	"noelle/src/types"
)

// Holds the Opencode session ID and metadata for a thread
type SessionEntry struct {
	SessionID		string
	Mode			types.BotMode
	LastActivity	time.Time
}

// Provides thread-safe access to session mappings
type SessionStore struct {
	mu			sync.RWMutex
	sessions	map[string]SessionEntry // Key is the Slack ThreadTS
}

// Initializes and returns a ready-to-use store
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]SessionEntry),
	}
}

// Safely retrieve a session entry for a given Slack thread
func (s *SessionStore) Get(threadTS string) (SessionEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.sessions[threadTS]
	return entry, exists
}

// Safely stores or updates a sessionentry for a given Slack thread
func (s *SessionStore) Set(threadTS string, entry SessionEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[threadTS] = entry
}

// Update the LastActivity timestamp of an existing session
func (s *SessionStore) Touch(threadTS string) {
	s.mu.Lock()
	defer s.mu.Lock()

	if entry, exists := s.sessions[threadTS]; exists {
		entry.LastActivity = time.Now()
		s.sessions[threadTS] = entry
	}
}

// Removes sessions that have been idle longer than maxAge
// This prevents memory from growing indefinitely over weeks of use
func (s *SessionStore) Cleanup(maxAge time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	evictedCount := 0

	for threadTS, entry := range s.sessions {
		if entry.LastActivity.Before(cutoff) {
			delete(s.sessions, threadTS)
			evictedCount++
		}
	}

	return evictedCount
}