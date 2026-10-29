package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"noelle/src/types"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// Runs a single worker loop. It should be launched as a goroutine:
// go StartWorker(1, ocService, slackClient, jobs)
func StartWorker(id int, ocService *OpenCodeService, slackClient *socketmode.Client, jobs <-chan types.QueryJob) {
	slog.Info(fmt.Sprintf("Worker %d started and ready for jobs.", id))

	for job := range jobs {
		slog.Info(fmt.Sprintf("Worker %d picked up query for thread %s [mode: %s]", id, job.ThreadTS, job.Mode.String()))

		// Set a timeout of 5 minutes for the OpenCode call
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

		// 1. Get or create the session
		sessionID, err := ocService.GetOrCreateSession(ctx, job)
		if err != nil {
			slog.Error(fmt.Sprintf("Worker %d failed to obtain session: %v", id, err))
			sendSlackReply(slackClient, job.ChannelID, job.ThreadTS, "Sorry, I ran into an issue connecting to the LLM session.")
			cancel()
			continue
		}

		// 2. Query OpenCode
		reply, err := ocService.Query(ctx, sessionID, job)
		cancel() // Release context resources as soon as the query returns

		if err != nil {
			slog.Error(fmt.Sprintf("Worker %d query failed: %v", id, err))
			sendSlackReply(slackClient, job.ChannelID, job.ThreadTS, "Sorry, an error occurred while processing your request.")
			continue
		}

		// 3. Post reply back to Slack thread
		sendSlackReply(slackClient, job.ChannelID, job.ThreadTS, reply)
		slog.Info(fmt.Sprintf("Worker %d finished query for thread %s", id, job.ThreadTS))
	}

	slog.Info(fmt.Sprintf("Worker %d stopped.", id))
}

func sendSlackReply(client *socketmode.Client, channelID, threadTS, text string) {
	_, _, err := client.PostMessage(
		channelID,
		slack.MsgOptionText(text, false),
		slack.MsgOptionTS(threadTS), // Ties the reply into the thread
	)

	if err != nil {
		slog.Error(fmt.Sprintf("Failed to post message to Slack channel %s: %v", channelID, err))
	}
}
