package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"github.com/sst/opencode-sdk-go"
	"github.com/sst/opencode-sdk-go/option"
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
	llmServerIP := os.Getenv("OPENCODE_SERVER_IP")
	llmServerPort := os.Getenv("OPENCODE_SERVER_PORT")
	llmServerPass := os.Getenv("OPENCODE_SERVER_PASSWORD")


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

	if llmServerIP == "" {
		fmt.Fprintf(os.Stderr, "OPENCODE_SERVER_IP environment variable is required.\n")
		os.Exit(1)
	}

	if llmServerPort == "" {
		fmt.Fprintf(os.Stderr, "OPENCODE_SERVER_PORT environment variable is required.\n")
		os.Exit(1)
	}

	if llmServerPort == "" {
		fmt.Fprintf(os.Stderr, "OPENCODE_SERVER_PORT environment variable is required.\n")
		os.Exit(1)
	}

	if llmServerPass == "" {
		fmt.Fprintf(os.Stderr, "OPENCODE_SERVER_PASSWORD environment variable is required.\n")
		os.Exit(1)
	}

	// Connect to Slack API and start a client
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

	// Connect to Opencode server
	authStr := fmt.Sprintf("opencode:%s", llmServerPass)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(authStr))
	header := fmt.Sprintf("Basic %s", encodedAuth)

	url := fmt.Sprintf("http://%s:%s", llmServerIP, llmServerPort)

	client := opencode.NewClient(
		option.WithBaseURL(url),
		option.WithHeader("Authorization", header),
		option.WithDebugLog(log.New(os.Stdout, "Opencode Client: ", log.Lshortfile|log.LstdFlags)),
	)

	// sessions, err := client.Session.List(context.TODO(), opencode.SessionListParams{})

	// if err != nil {
	// 	fmt.Printf("Failed to get sessions from the Opencode server. Info:\n%s", err)
	// 	os.Exit(1)
	// }

	fmt.Printf("Connected successfully! Active sessions:\n%+v\n", sessions)
	
	go RunMonitor(slackClient)
	go RunQuery(&llmServerIP, &llmServerPort, &llmServerPass)	

	slackClient.Run()
}