package main

import (
	"context"
	"fmt"
	"log/slog"
	"noelle/src/types"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// StartWorker runs a single worker loop. It should be launched as a goroutine:
//
// go StartWorker(1, ocService, slackClient, jobs)
func StartWorker(id int, ocService *OpenCodeService, slackClient *socketmode.Client, jobs <-chan types.QueryJob) {
	slog.Info(fmt.Sprintf("Worker %d started and ready for jobs.", id))

	for job := range jobs {
		slog.Info(fmt.Sprintf("Worker %d picked up query for thread %s [mode: %s]", id, job.ThreadTS, job.Mode.String()))

		// 1. Post an initial placeholder message in the Slack thread
		_, msgTs, err := slackClient.PostMessage(
			job.ChannelID,
			slack.MsgOptionText("Processing your request...", false),
			slack.MsgOptionTS(job.ThreadTS),
		)

		if err != nil {
			slog.Error(fmt.Sprintf("Worker %d failed to post initial message: %v", id, err))
			continue
		}

		// 2. Start a background goroutine to rotate status messages every 2 seconds
		stopStatus := make(chan struct{})
		go func(channelId, targetTS string) {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			phrases := []string{
				"Researching project files...",
				"Analyzing project structure...",
				"Thinking...",
				"Tailoring explanations...",
				"This is taking a while...",
				"Still working on it...",
				"Optimizing...",
			}
			idx := 0

			for {
				select {
				case <-stopStatus:
					// Exit the animation loop when signaled
					return
				case <-ticker.C:
					currentText := phrases[idx%len(phrases)]
					idx++

					_, _, _, updErr := slackClient.UpdateMessage(
						channelId,
						targetTS,
						slack.MsgOptionText(currentText, false),
					)

					if updErr != nil {
						slog.Debug("Status update tick failed: %v", updErr)
					}
				}
			}
		}(job.ChannelID, msgTs)

		// 3. Obtain session and execute the query
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

		sessionID, err := ocService.GetOrCreateSession(ctx, job)
		if err != nil {
			close(stopStatus)
			cancel()
			slog.Error(fmt.Sprintf("Worker %d failed to obtain session: %v", id, err))
			_, _, _, _ = slackClient.UpdateMessage(
				job.ChannelID,
				msgTs,
				slack.MsgOptionText("Sorry, I ran into an issue connecting to the LLM session.", false),
			)
			continue
		}

		// Query OpenCode with event streaming
		reply, err := ocService.Query(ctx, sessionID, job)
		close(stopStatus) // Stop status ticker since we have the full answer now
		cancel()

		if err != nil {
			slog.Error(fmt.Sprintf("Worker %d failed to query OpenCode: %v", id, err))
			_, _, _, _ = slackClient.UpdateMessage(
				job.ChannelID,
				msgTs,
				slack.MsgOptionText("Sorry, an error occurred while processing your request.", false),
			)
			continue
		}

		// 4. Overwrite the placeholder message with the final answer

		// The current limit of Slack messages is 4000 runes, so we must ensure it will not exceed
		msgArray := SplitSlackMessage(reply, 4000)

		// Update the original message with the first part of the reply
		// and post in new messages as required the rest of the response
		_, _, _, err = slackClient.UpdateMessage(
			job.ChannelID,
			msgTs,
			slack.MsgOptionText(msgArray[0], false),
		)

		if err != nil {
			slog.Error(fmt.Sprintf("Worker %d failed to update final message: %v", id, err))
		}

		if err != nil && len(msgArray) > 1 {
			for i, msg := range msgArray[1:] {
				_, _, err := slackClient.PostMessage(
					job.ChannelID,
					slack.MsgOptionText(msg, false),
					slack.MsgOptionTS(job.ThreadTS),
				)

				if err != nil {
					slog.Error(fmt.Sprintf("Worker %d failed to post part %d message: %v\nStopping response thread to prevent inconsistent replies!", id, i, err))
					break
				} else {
					time.Sleep(250 * time.Millisecond) // Sleep to prevent API rate limiting
				}
			}
		}

		slog.Info(fmt.Sprintf("Worker %d finished query for thread %s.", id, job.ThreadTS))
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
