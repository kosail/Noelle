package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"noelle/src/types"

	"github.com/sst/opencode-sdk-go"
)

// Wraps the SDK client and provides high-level helpers.
type OpenCodeService struct {
	client *opencode.Client
	store  *SessionStore
}

// Creates a new service wrapper
func NewOpenCodeService(client *opencode.Client, store *SessionStore) *OpenCodeService {
	return &OpenCodeService{
		client: client,
		store:  store,
	}
}

// Retrieves the session for a thread, or creates a new one
func (s *OpenCodeService) GetOrCreateSession(ctx context.Context, job types.QueryJob) (string, error) {
	// 1. Check if we already have a session for this Slack thread
	if entry, exists := s.store.Get(job.ThreadTS); exists {
		s.store.Touch(job.ThreadTS)
		return entry.SessionID, nil
	}

	// 2. If not, create a new session in Opencode and store it into the store
	title := fmt.Sprintf("Slack Thread: %s (%s)", job.ThreadTS, job.Mode.String())
	newSession, err := s.client.Session.New(ctx, opencode.SessionNewParams{
		Title: opencode.F(title),
	})
	if err != nil {
		return "", fmt.Errorf("Failed creating Opencode session: %w", err)
	}

	s.store.Set(job.ThreadTS, SessionEntry{
		SessionID:    newSession.ID,
		Mode:         job.Mode,
		LastActivity: time.Now(),
	})

	return newSession.ID, nil
}

// Send the user question to the session and extracts the answer
func (s *OpenCodeService) Query(ctx context.Context, sessionID string, job types.QueryJob) (string, error) {
	//Construct the prompt
	parts := []opencode.SessionPromptParamsPartUnion{
		opencode.TextPartInputParam{
			Type: opencode.F(opencode.TextPartInputTypeText),
			Text: opencode.F(job.Prompt),
		},
	}

	// Send the prompt to Opencode
	resp, err := s.client.Session.Prompt(ctx, sessionID, opencode.SessionPromptParams{
		Parts:  opencode.F(parts),
		System: opencode.F(job.Mode.SystemPrompt()),
	})

	if err != nil {
		return "", fmt.Errorf("Opencode prompt error: %w", err)
	}

	// Extract the text parts from the response
	var answer strings.Builder
	for _, part := range resp.Parts {
		if part.Text != "" {
			answer.WriteString(part.Text)
		}
	}

	result := strings.TrimSpace(answer.String())
	if result == "" {
		// Simple fallback
		return "I reviewed the files, but did not find an answe for that query", nil
	}

	return result, nil
}
