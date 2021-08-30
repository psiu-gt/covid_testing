package main

import (
	"github.com/juju/errors"
	"github.com/slack-go/slack"
)

type Slack struct {
	client    *slack.Client
	channelID string
}

// New initializes the Slack object.
func (s *Slack) New(apiKey, channelID string) {
	s.client = slack.New(apiKey)
	s.channelID = channelID
}

func (s *Slack) SendMessage(message string) error {
	_, _, err := s.client.PostMessage(s.channelID, slack.MsgOptionText(message, false))
	if err != nil {
		return errors.Annotate(err, "Slack.SendMessage(): failed to post message")
	}

	return nil
}
