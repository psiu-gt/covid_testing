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
	s.client = slack.New(apiKey, slack.OptionDebug(true))
	s.channelID = channelID
}

// Send a message to Slack.
func (s *Slack) SendMessage(message string) error {
	_, _, err := s.client.PostMessage(s.channelID, slack.MsgOptionText(message, false))
	if err != nil {
		return errors.Annotate(err, "client.PostMessage(): failed to post message")
	}

	return nil
}

// Get a list of all user IDs.
func (s *Slack) GetUsers() ([]string, error) {
	users, _, err := s.client.GetUsersInConversation(&slack.GetUsersInConversationParameters{ChannelID: s.channelID})
	if err != nil {
		return nil, errors.Annotate(err, "client.GetUsersInConversation(): failed to get users")
	}

	// Filter out the bot.
	filtered := make([]string, 0)
	for _, user := range users {
		// ID of the bot itself.
		// TODO(dwitt): Not hardcode this value.
		if user != "U02D7EJ134Z" {
			filtered = append(filtered, user)
		}
	}

	return filtered, nil
}

// Get the Name for each user id.
func (s *Slack) GetUserRealNames(userIDs []string) ([]string, error) {
	resp, err := s.client.GetUsersInfo(userIDs...)
	if err != nil {
		return nil, errors.Annotate(err, "GetUserNames(): failed to get user info from Slack")
	}

	names := make([]string, len(*resp))
	for i, user := range *resp {
		names[i] = user.RealName
	}

	return names, nil
}
