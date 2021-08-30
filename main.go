// Package main provides scraping and alerting for testing.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

type cfg struct {
	SheetsID       string `json:"sheetsID"`
	SlackToken     string `json:"slackToken"`
	SlackChannelID string `json:"slackChannelID"`
}

type TestResult struct {
	Name           string
	WithWeek       bool
	TestDate       string
	LastTestResult string
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

func main() {
	config, err := readConfigFromFile("config.json")
	if err != nil {
		log.Fatalf("readConfigFromFile: %v", err)
	}

	sheets := Sheets{sheetID: config.SheetsID}
	results, err := sheets.ReadSheets()
	if err != nil {
		log.Fatalf("sheets.ReadSheets(): %v", err)
	}

	slackClient := Slack{}
	slackClient.New(config.SlackToken, config.SlackChannelID)
	err = slackClient.SendMessage(fmt.Sprintf("%v", results))
	if err != nil {
		log.Fatalf("slackClient.SendMessage(): %v", err)
	}

	fmt.Println(results)
}
