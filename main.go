// Package main provides scraping and alerting for testing.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

type cfg struct {
	SheetsID        string `json:"sheetsID"`
	SlackToken      string `json:"slackToken"`
	SlackChannelID  string `json:"slackChannelID"`
	SheetReadRange  string `json:"sheetReadRange"`
	SheetWriteRange string `json:"sheetWriteRange"`
}

func readConfigFromFile(filename string) (*cfg, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := cfg{}
	json.Unmarshal(file, &config)

	return &config, nil
}

func constructNotificationMessage(untested *[]TestResult, nameToIDs map[string]string) string {
	if len(*untested) == 0 {
		fmt.Println("No untested names, returning.")
		return ""
	}

	lines := make([]string, 0)
	lines = append(lines, "The following people have not submitted a test result in the past week:")
	for _, entry := range *untested {
		id, ok := nameToIDs[entry.Name]
		if !ok {
			fmt.Println("WARNING: entry not found, skipping", entry.Name)
			continue
		}
		lines = append(lines, fmt.Sprintf("<@%s>", id))
	}
	lines = append(lines, "GT Surveillance Testing Locations & Hours: https://health.gatech.edu/coronavirus/testing")
	lines = append(lines, "Form Link: https://forms.gle/sydSGQpTEgPrGxWy9")
	return strings.Join(lines, "\n")
}

func main() {
	log.Println("Starting service...")

	config, err := readConfigFromFile("config.json")
	if err != nil {
		log.Fatalf("readConfigFromFile: %v", err)
	}

	// Create the Google Sheets client.
	sheets := Sheets{}
	err = sheets.New(config.SheetsID, config.SheetReadRange, config.SheetWriteRange)
	if err != nil {
		log.Fatalf("sheets.New(): %v", err)
	}

	// Create the Slack client.
	slackClient := Slack{}
	slackClient.New(config.SlackToken, config.SlackChannelID)

	// Update the list of names for the Google Form.
	// The names are extracted from Slack and written to Google Sheets, which populates the Google Forms dropdown.
	log.Println("Updating Google Sheets with Slack users.")
	users, err := slackClient.GetUsers()
	if err != nil {
		log.Fatalf("slackClient.GetUsers(): %v", err)
	}
	names, err := slackClient.GetUserRealNames(users)
	if err != nil {
		log.Fatalf("slackClient.GetUserRealNames(): %v", err)
	}
	err = sheets.WriteNames(names)
	if err != nil {
		log.Fatalf("sheets.WriteNames(): %v", err)
	}

	// Construct a map of names to IDs (for mentions in Slack)
	namesToID := make(map[string]string)
	for i, name := range names {
		namesToID[name] = users[i]
	}

	// Get the test results.
	log.Println("Getting test results from Google Sheets...")
	results, err := sheets.ReadSheets()
	if err != nil {
		log.Fatalf("sheets.ReadSheets(): %v", err)
	}

	// Filter for the people to scream (gently) at.
	untested := GetUntested(results)

	msg := constructNotificationMessage(untested, namesToID)
	if msg == "" {
		log.Fatalf("no message to send")
	}

	log.Println("Sending message to Slack")
	err = slackClient.SendMessage(msg)
	if err != nil {
		log.Fatalf("slackClient.SendMessage(): %v", err)
	}
}
