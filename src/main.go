package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func main() {
	// Check if .env is correctly set up
	err := godotenv.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading .env file!")
		os.Exit(1)
	}

	botToken := os.Getenv("SLACK_BOT_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")
	llmToken := os.Getenv("LLM_API_TOKEN")
	llmModel := os.Getenv("LLM_MODEL")


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

	if llmToken == "" {
		fmt.Fprintf(os.Stderr, "LLM_API_TOKEN environment variable is required.\n")
		os.Exit(1)
	}

	if !strings.HasPrefix(llmToken, "oc_sk_") {
		fmt.Fprintf(os.Stderr, "LLM_API_TOKEN must have the prefix \"oc_sk_\", and be a valid OpenCode token.\n")
		os.Exit(1)
	}

	if llmModel == "" {
		fmt.Fprintf(os.Stderr, "LLM_MODEL environment variable is required.\n")
		os.Exit(1)
	}

	// Note: This is a simple program, so it does not check if the typed LLM model is valid against
	// the actual LLM list that OpenCode offers. Maybe in the future tho...

	slackApi := slack.New(
		botToken,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "Slack API: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(appToken),
	)

	slackClient := socketmode.New(
		slackApi,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "Slack Socket: ", log.Lshortfile|log.LstdFlags)),
	)
	
	go StartBot(slackClient)

	slackClient.Run()
}