package types

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// BotMode represents which personality or focus noelle will use
type BotMode int

const (
	ModeNormal BotMode = iota
	ModeArtist
	ModeDev
)

// Return a friendly name for the logger
func (m BotMode) String() string {
	switch m {
	case ModeArtist:
		return "Artist Mode"
	case ModeDev:
		return "Dev Mode"
	default:
		return "Normal Mode"
	}
}

// ContextFileName returns the filename corresponding to this mode
// inside the context/ folder
func (m BotMode) ContextFileName() string {
	switch m {
	case ModeArtist:
		return "artist.md"
	case ModeDev:
		return "dev.md"
	default:
		return "normal.md"
	}
}

// SystemPromptreads the mode's instructions from the context/ folder on disk
// If the file is missing or cannot be read, it falls back to a built-in default.
func (m BotMode) SystemPrompt() string {
	filePath := filepath.Join("context", m.ContextFileName())

	data, err := os.ReadFile(filePath)
	if err == nil {
		prompt := strings.TrimSpace(string(data))
		if prompt != "" {
			return prompt
		}
	}

	// File was missing, empty or unreadable: log notice and use fallback
	slog.Warn("Could not read prompt file, using built-in fallback", "file", filePath, "error, err")
	return m.defaultSystemPrompt()
}

// defaultSystemPrompt provides safe in-memory defaults if files are not yet created.
func (m BotMode) defaultSystemPrompt() string {
	switch m {
	case ModeArtist:
		return "You are an assistant for Unity game artists and gameplay mechanics developers. " +
			"Focus on Unity components, Inspector parameters, practical mechanics, and simple C# scripting. " +
			"Avoid deep software architecture, infrastructure details, and unnecessary backend jargon. " +
			"You are strictly a read-only assistant; never attempt to modify or create files."

	case ModeDev:
		return "You are an assistant for engine and core systems developers. " +
			"Provide in-depth architectural analysis, memory and performance considerations, " +
			"low-level code references, and systems design. " +
			"You are strictly a read-only assistant; never attempt to modify or create files."

	default:
		return "You are a helpful project assistant. Provide clear, balanced answers " +
			"about the project files and documentation. " +
			"You are strictly a read-only assistant; never attempt to modify or create files."
	}
}

// represents a single inquiry waiting to be processed by a worker
type QueryJob struct {
	ChannelID string // Slack channel where the question was asked
	ThreadTS  string // Slack thread timestamp (used for goroutines threading)
	UserID    string // Slack user ID whom made the question
	Prompt    string // Cleaned text of the question
	Mode      BotMode
}
