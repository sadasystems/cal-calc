package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"

	"github.com/jinzhu/now"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
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
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	min := now.BeginningOfWeek()
	max := now.EndOfWeek()

	events, err := srv.
		Events.List("primary").
		ShowDeleted(false).
		OrderBy("startTime").
		SingleEvents(true).
		TimeMin(min.Format(time.RFC3339)).
		TimeMax(max.Format(time.RFC3339)).
		Do()
	if err != nil {
		log.Fatalf("Unable to retrieve events from the past week: %v", err)
	}

	fmt.Printf("Attended events between %s and %s:\n\n", min.Format(time.RFC850), max.Format(time.RFC850))

	if len(events.Items) == 0 {
		fmt.Println("No events found.")
	} else {
		for _, item := range events.Items {
			// filter out events we didn't accept
			attended := true
			for _, attendee := range item.Attendees {
				if attendee.Self && attendee.ResponseStatus != "accepted" {
					attended = false
				}
			}

			// filter out all-day events
			allday := false
			startDate := item.Start.DateTime
			endDate := item.End.DateTime

			if startDate == "" || item.EventType == "outOfOffice" {
				// the only all-day events we care about are OOO
				// but those already mark affected events as declined, so ignore them
				// also it turns out OOO aren't all-day events
				allday = true
			}

			if attended && !allday {
				fmt.Printf("%v (%v to %v)\n", item.Summary, startDate, endDate)
			}
		}
	}
}
