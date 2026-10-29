package main

import (
	"fmt"
	"log/slog"
	"noelle/src/types"
	"regexp"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// Regex to strip Slack user mentions like <@U12345678> from text
var mentionRegex = regexp.MustCompile(`<@[A-Z0-9]+>`)

// Listen for incoming events from Slack's WebSocket connection
func RunMonitor(client *socketmode.Client, jobs chan<- types.QueryJob) {
	for evt := range client.Events {
		switch evt.Type {
		case socketmode.EventTypeConnecting:
			slog.Info("Connecting to Slack with Socket Mode...")
		case socketmode.EventTypeConnectionError:
			slog.Error("Connection failed. Retrying later...")
		case socketmode.EventTypeConnected:
			slog.Info("Connected to Slack with Socket Mode.")
		case socketmode.EventTypeEventsAPI:
			handleEventsAPI(client, evt, jobs)

		case socketmode.EventTypeSlashCommand:
			_, ok := evt.Data.(slack.SlashCommand)
			if !ok {
				msg := fmt.Sprintf("Ignored event: %v\n", evt)
				slog.Debug(msg)
				continue
			}

		}
	}
}

// Process @Noelle mentions in channels and threads.
func handleEventsAPI(client *socketmode.Client, evt socketmode.Event, jobs chan<- types.QueryJob) {
	event, ok := evt.Data.(slackevents.EventsAPIEvent)

	if !ok {
		slog.Debug(fmt.Sprintf("Ignored event: %+v\n", event))
		return
	}

	// Quick acknowledge the event so Slack knows we received it
	client.Ack(*evt.Request)

	if event.Type != string(slackevents.CallbackEvent) {
		return
	}

	switch ev := event.InnerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		// Sanitize the prompt removing the `@` user mention and trimming
		cleanPrompt := strings.TrimSpace(mentionRegex.ReplaceAllString(ev.Text, ""))
		if cleanPrompt == "" {
			sendSlackReply(client, ev.Channel, ev.TimeStamp, "Hello! Ask me any question about the project files, structure or any guidance you need.")
			return
		}

		// If asked inside an existing thread, reuse that thread timestamp.
		//Otherwise, use this message's own timestamp as the new thread root.
		threadTS := ev.ThreadTimeStamp
		if threadTS == "" {
			threadTS = ev.TimeStamp
		}

		// Push the job to the workers
		jobs <- types.QueryJob{
			ChannelID: ev.Channel,
			ThreadTS: threadTS,
			UserID: ev.User,
			Prompt: cleanPrompt,
			Mode: types.ModeNormal,
		}
	}
}

// Process /artist and /dev commands.
func handleSlashCommand(client *socketmode.Client, evt socketmode.Event, jobs chan<- types.QueryJob) {
	
}
