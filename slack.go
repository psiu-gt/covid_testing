package main

import (
	"github.com/slack-go/slack"
)

type Slack struct {
	client *slack.Client
}

// New initializes the Slack object.
func (s *Slack) New(apiKey string) {
	s.client = slack.New(apiKey)
}
