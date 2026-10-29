package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sst/opencode-sdk-go"
	"noelle/src/types"
)

// Wraps the SDK client and provides high-level helpers.
type OpenCodeService struct {
	client	*opencode.Client
	store	*SessionStore
}

// Creates a new service wrapper
func NewOpenCodeService(client *opencode.Client, store *SessionStore) *OpenCodeService {
	return &OpenCodeService{
		client: client,
		store: store,
	}
}

// Retrieves the session for a thread, or creates a new one
func (s *OpenCodeService) GetOrCreateSession(ctx context.Context)