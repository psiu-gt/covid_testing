package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/juju/errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

type Sheets struct {
	sheetID string
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) (*http.Client, error) {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok, err = getTokenFromWeb(config)
		if err != nil {
			return nil, errors.Annotate(err, "getClient(): failed to get the token from web")
		}
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok), nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, errors.Annotate(err, "getTokenFromWeb(): unable to read authorization code")
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, errors.Annotate(err, "getTokenFromWeb(): unable to retrieve token from web")
	}
	return tok, nil
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Annotate(err, "saveToken(): unable to cache oauth token")
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)

	return nil
}

// ReadSheets reads the Google Sheet and returns the data.
func (s Sheets) ReadSheets() (*[]TestResult, error) {
	// Read the credentials.
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		return nil, errors.Annotate(err, "ReadSheets(): unable to readclient secret file credentials.json")
	}

	// If modifying these scopes, delete the previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		return nil, errors.Annotate(err, "ReadSheets(): unable to parse client secret file to config")
	}
	client, err := getClient(config)
	if err != nil {
		return nil, errors.Annotate(err, "ReadSheets(): unable to create the client")
	}

	// Create the client with the obtained credentials.
	srv, err := sheets.New(client)
	if err != nil {
		return nil, errors.Annotate(err, "ReadSheets(): unable to retrieve Google Sheets client")
	}

	// Prints the names and majors of students in a sample spreadsheet:
	// https://docs.google.com/spreadsheets/d/1ordb1uYNAvJHnQAN_7uLjG0d8gUhPSvjfQDQ7XHGrq8/edit#gid=1642678966
	readRange := "Dashboard!A4:D"
	resp, err := srv.Spreadsheets.Values.Get(s.sheetID, readRange).Do()
	if err != nil {
		return nil, errors.Annotate(err, "ReadSheets(): unable to retrieve data from sheet")
	}

	if len(resp.Values) == 0 {
		return nil, errors.Annotate(err, "ReadSheets(): no data in sheet")
	}

	var results []TestResult = make([]TestResult, len(resp.Values))

	for i, row := range resp.Values {
		if len(row) != 0 {
			results[i].Name = row[0].(string)
			results[i].WithWeek = row[1] == "TRUE"
			results[i].TestDate = row[2].(string)
			results[i].LastTestResult = row[3].(string)
		}
	}

	return &results, nil
}
