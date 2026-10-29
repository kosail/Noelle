package main

import (
	"fmt"
	"log/slog"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func RunMonitor(client *socketmode.Client, comm chan string) {
	for evt := range client.Events {
		switch evt.Type {
			case socketmode.EventTypeConnecting:
				slog.Info("Connecting to Slack with Socket Mode...")
			case socketmode.EventTypeConnectionError:
				slog.Error("Connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				slog.Info("Connected to Slack with Socket Mode.")
			case socketmode.EventTypeEventsAPI:
				event, ok := evt.Data.(slackevents.EventsAPIEvent)

				if !ok {
					err := fmt.Sprintf("Ignored event: %+v\n", event)
					slog.Debug(err)
					continue
				}

				client.Ack(*evt.Request)

				if event.Type != string(slackevents.CallbackEvent) {
					continue
				}

				switch ev := event.InnerEvent.Data.(type) {
				case *slackevents.AppMentionEvent:
					msg := fmt.Sprintf("Message from %s: %s\n", ev.User, ev.Text)
					_, _, err := client.PostMessage(ev.Channel, slack.MsgOptionText(msg, false))

					if err != nil {
						msg := fmt.Sprintf("Failed posting message: %v", err)
						slog.Error(msg)
					}

					comm <- ev.Text
				}

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