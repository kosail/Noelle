package types

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

// return the specialized instruction set for OpenCode
// TODO: Change this into a proper AGENTS.md for each use-case.
// As per now, let's use simple prompts to ensure first it works
func (m BotMode) SystemPrompt() string {
	switch m {
	case ModeArtist:
		return "You are an assistant for Unity game artists and gameplay mechanics developers. " +
			"Focus on Unity components, Inspector parameters, practical mechanics, and simple C# scripting. " +
			"Avoid deep software architecture, infrastructure details, and unnecessary backend jargon."
	case ModeDev:
		return "You are an assistant for engine and core systems developers. " +
			"Provide in-depth architectural analysis, memory and performance considerations, " +
			"low-level code references, and systems design."
	default:
		return "You are a helpful project assistant. Provide clear, balanced answers " +
			"about the project files and documentation."
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
