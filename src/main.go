package main

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"noelle/src/types"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"github.com/sst/opencode-sdk-go"
	"github.com/sst/opencode-sdk-go/option"
)

// Number of concurrent workers that will be taking requests
const numWorkers = 3

func main() {
	// Check if .env is correctly set up, and load env variables
	err := godotenv.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Notice: No .env file found or failed to load. Using system environment variables.\n")
	}

	botToken := os.Getenv("SLACK_BOT_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")
	ocServerIP := os.Getenv("OPENCODE_SERVER_IP")
	ocServerPort := os.Getenv("OPENCODE_SERVER_PORT")
	ocServerPass := os.Getenv("OPENCODE_SERVER_PASSWORD")
	ocProvider := os.Getenv("OPENCODE_PROVIDER")
	ocModel := os.Getenv("OPENCODE_MODEL")

	if appToken == "" {
		fmt.Fprintf(os.Stderr, "SLACK_APP_TOKEN environment variable is required\n")
		os.Exit(1)
	}

	if !strings.HasPrefix(appToken, "xapp-") {
		fmt.Fprintf(os.Stderr, "SLACK_APP_TOKEN must have the prefix \"xapp-\"\n")
		os.Exit(1)
	}

	if botToken == "" {
		fmt.Fprintf(os.Stderr, "SLACK_BOT_TOKEN environment variable is required\n")
		os.Exit(1)
	}

	if !strings.HasPrefix(botToken, "xoxb-") {
		fmt.Fprintf(os.Stderr, "SLACK_BOT_TOKEN must have the prefix \"xoxb-\"\n")
		os.Exit(1)
	}

	if ocServerIP == "" || ocServerPort == "" {
		fmt.Fprintf(os.Stderr, "OPENCODE_SERVER_IP and OPENCODE_SERVER_PORT environment variables are required.\n")
		os.Exit(1)
	}

	if ocServerPass == "" {
		fmt.Fprintf(os.Stderr, "OPENCODE_SERVER_PASSWORD environment variable is required.\n")
		os.Exit(1)
	}

	if ocProvider == "" || ocModel == "" {
		slog.Warn("OPENCODE_PROVIDER or OPENCODE_MODEL not found. Using OpenCode's server default config")
	} else {
		slog.Info(fmt.Sprintf("Using configured OpenCode model: %s/%s", ocProvider, ocModel))
	}

	// 1.5 Check if the context directory exists and can be located
	contextDir := "context"
	if info, err := os.Stat(contextDir); err == nil && info.IsDir() {
		slog.Info("Found context/ folder for external prompts.")
	} else {
		slog.Warn("context/ folder not found. Noelle will use built-in fallback prompts.", "path", contextDir)
	}

	// 2. Connect to Slack API and start a socketmode client
	slackApi := slack.New(
		botToken,
		slack.OptionAppLevelToken(appToken),
	)

	slackClient := socketmode.New(
		slackApi,
	)

	// 3. Connect to OpenCode server
	authStr := fmt.Sprintf("opencode:%s", ocServerPass)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(authStr))
	authHeader := fmt.Sprintf("Basic %s", encodedAuth)
	baseUrl := fmt.Sprintf("http://%s:%s", ocServerIP, ocServerPort)

	ocClient := opencode.NewClient(
		option.WithBaseURL(baseUrl),
		option.WithHeader("Authorization", authHeader),
	)

	// 4. Initialize the session store and OC service
	sessionStore := NewSessionStore()
	ocService := NewOpenCodeService(ocClient, sessionStore, ocProvider, ocModel)

	// 5. Create the buffered job channel (by now 30 items cap)
	jobQueue := make(chan types.QueryJob, 30)

	// 6. Launch background workers to handle queries concurrently
	for i := 1; i <= numWorkers; i++ {
		go StartWorker(i, ocService, slackClient, jobQueue)
	}

	// 7. Start a background ticker to clean up idle sessions every 30 min
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		for range ticker.C {
			evicted := sessionStore.Cleanup(1 * time.Hour)
			if evicted > 0 {
				slog.Info(fmt.Sprintf("Cleaned up %d idle OpenCode sessions from memory.", evicted))
			}
		}
	}()

	// 8. Start the Slack event monitor
	go RunMonitor(slackClient, jobQueue)
	slog.Info("Noelle bot is running and listening to Slack!")

	// 9. Run the slack client
	// (this already blocks the main goroutine, so Noelle stays alive)
	err = slackClient.Run()
	if err != nil {
		slog.Error("Slack Socker Mode encountered an error: %v", err)
		os.Exit(1)
	}
}
