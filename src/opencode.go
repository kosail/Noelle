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
	client   *opencode.Client
	store    *SessionStore
	provider string
	model    string
}

// Creates a new service wrapper
func NewOpenCodeService(client *opencode.Client, store *SessionStore, provider, model string) *OpenCodeService {
	return &OpenCodeService{
		client:   client,
		store:    store,
		provider: provider,
		model:    model,
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
	// 1. Open the event stream BEFORE the prompt so we don't miss early events
	stream := s.client.Event.ListStreaming(ctx, opencode.EventListParams{})
	defer stream.Close()

	// 2. Construct the prompt
	parts := []opencode.SessionPromptParamsPartUnion{
		opencode.TextPartInputParam{
			Type: opencode.F(opencode.TextPartInputTypeText),
			Text: opencode.F(job.Prompt),
		},
	}

	// Setting up the prompt mode. This bot will ever be read-only.
	// Granting write permissions would be begging for a nuke in the server. Just, don't.
	promptParams := opencode.SessionPromptParams{
		Parts:  opencode.F(parts),
		System: opencode.F(job.Mode.SystemPrompt()),
		Agent:  opencode.F("plan"), // Enforce read-only mode
		Tools: opencode.F(map[string]bool{
			"edit":  false,
			"write": false,
			"patch": false,
		}),
	}

	// Attach model if configuredor use defaults of the OC server
	if s.provider != "" && s.model != "" {
		promptParams.Model = opencode.F(opencode.SessionPromptParamsModel{
			ProviderID: opencode.F(s.provider),
			ModelID:    opencode.F(s.model),
		})
	}

	// 3. Dispatch the prompt to OpenCode
	_, err := s.client.Session.Prompt(ctx, sessionID, promptParams)
	if err != nil {
		return "", fmt.Errorf("opencode prompt error: %w", err)
	}

	// 4.Listen to the stream until OpenCode finishes (session.idle becomes true)
	var answer strings.Builder

	for stream.Next() {
		evt := stream.Current()

		switch u := evt.AsUnion().(type) {
		case opencode.EventListResponseEventMessagePartUpdated:
			part := u.Properties.Part

			// Make sure if the event belongs to our thread's session
			if part.SessionID != sessionID {
				continue
			}

			// Only collect text parts, ignore reasoning
			if part.Type == opencode.PartTypeText {
				if u.Properties.Delta != "" {
					answer.WriteString(u.Properties.Delta)
				}
			}

		case opencode.EventListResponseEventSessionIdle:
			// When our session becomes idle, all thinking and tools are finished
			if u.Properties.SessionID == sessionID {
				result := strings.TrimSpace(answer.String())
				if result == "" {
					return "I reviewed the files, but did not find an answer for that query.", nil
				}
				return result, nil
			}
		}
	}
	// If the stream closed or context expired before session.idle
	if err := stream.Err(); err != nil {
		return "", fmt.Errorf("opencode stream error: %w", err)
	}

	finalResult := strings.TrimSpace(answer.String())
	if finalResult != "" {
		return finalResult, nil
	}

	return "I reviewed the files, but did not find an answer for that query.", nil
}
