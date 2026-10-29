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
			ThreadTS:  threadTS,
			UserID:    ev.User,
			Prompt:    cleanPrompt,
			Mode:      types.ModeNormal,
		}
	}
}

// Process /artist and /dev commands.
func handleSlashCommand(client *socketmode.Client, evt socketmode.Event, jobs chan<- types.QueryJob) {
	cmd, ok := evt.Data.(slack.SlashCommand)
	if !ok {
		slog.Debug(fmt.Sprintf("Ignored slash event: %+v\n", cmd))
		return
	}

	// Acknowledge immediately to Slack
	client.Ack(*evt.Request)

	cleanPrompt := strings.TrimSpace(cmd.Text)
	if cleanPrompt == "" {
		// If the user typed just the slash command without a question
		msg := fmt.Sprintf("Please provide a question after %s. Example: `%s How do the tagging system works for my X specific use case?`", cmd.Command, cmd.Command)
		_, _, _ = client.PostMessage(cmd.ChannelID, slack.MsgOptionText(msg, false))
		return
	}

	// Determine the BotMode based on the command name
	mode := types.ModeNormal
	switch cmd.Command {
	case "/artist":
		mode = types.ModeArtist
	case "/dev":
		mode = types.ModeDev
	}

	// Post an initial message in the channel to anchor the conversation
	announcement := fmt.Sprintf("<@%s> asked (%s): *%s*", cmd.UserID, mode.String(), cleanPrompt)
	_, rootTS, err := client.PostMessage(cmd.ChannelID, slack.MsgOptionText(announcement, false))

	if err != nil {
		slog.Error(fmt.Sprintf("Failed to post slash command announcement: %v", err))
		return
	}

	// Use rootTS so the worker will post the full answer inside the new thread
	jobs <- types.QueryJob{
		ChannelID: cmd.ChannelID,
		ThreadTS:  rootTS,
		UserID:    cmd.UserID,
		Prompt:    cleanPrompt,
		Mode:      mode,
	}
}
