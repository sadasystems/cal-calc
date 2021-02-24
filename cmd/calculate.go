package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"

	"github.com/jinzhu/now"
)

var debug = false
var info = true

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

func calculate() {
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

	// these are substrings we categorize meetings with
	meetingKeywords := map[string][]string{
		"Intellectual Property": []string{
			"IP",
		},
		"Internal Meetings": []string{
			"1:1 Slava:",
			"Town hall",
			"Anthos On-Prem Biweekly",
			"Braintrust",
			"Educational Webinar",
			"Benefits",
			"Sync w",
			"Beer",
			"InfraMod",
		},
		"OOO": []string{},
		"Billable": []string{
			"Visby",
			"Neustar",
			"N*",
			"CF",
			"EPS",
			"Governance",
		},
		"Management": []string{
			":Slava", // team 1:1s with me are set in the format "1:1 Employee:Manager"
			"Management",
			"Resourcing",
			"Team Building",
			"Werewolf",
			"Weekly G Cloud Ops",
		},
	}

	// a store for durations for each category
	allocations := map[string]time.Duration{
		"Intellectual Property": time.Duration(0),
		"Internal Meetings": time.Duration(0),
		"OOO": time.Duration(0),
		"Billable": time.Duration(0),
		"Management": time.Duration(0),
	}

	// unmatched meetings go here
	unmatched := []string{}

	timeFormat := "Monday, January 2"

	// total time found in meetings
	totalTime := time.Duration(0)

	fmt.Printf("Parsing events between %s and %s\n\n", min.Format(timeFormat), max.Format(timeFormat))

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
			if item.Start.DateTime == "" { continue }
			if item.EventType == "outOfOffice" {
				// the only all-day events we care about are OOO
				// but those already mark affected events as declined, so ignore those and just count this event alone
				// also it turns out OOO aren't all-day events

				allocations["OOO"] += 8 * time.Hour // TODO: who knows if this works right: one event per day, right?
				totalTime += 8 * time.Hour
				continue
			}

			start, _ := time.Parse(time.RFC3339, item.Start.DateTime)
			end, _ := time.Parse(time.RFC3339, item.End.DateTime)

			duration := end.Sub(start)

			if attended {
				if debug { fmt.Printf("%v %v\n", item.Summary, duration) }
				matched := false

				for meetingType, keywords := range meetingKeywords {
					for _, keyword := range keywords {
						if debug { fmt.Printf("  %v:%v\n", item.Summary, keyword) }
						if strings.Contains(strings.ToLower(item.Summary), strings.ToLower(keyword)) {
							if debug { fmt.Printf("    %v:%v\n", item.Summary, keyword) }
							if debug { fmt.Printf("%v assigned to %v\n", item.Summary, meetingType) }

							allocations[meetingType] += duration
							matched = true
						}
					}
				}

				if !matched {
					unmatched = append(unmatched, item.Summary)
				} else {
					totalTime += duration
				}
			}
		}
	}

	// print total time
	fmt.Printf("Total time: %v\n\n", totalTime)

	// print category totals
	for t, d := range allocations {
		fmt.Printf("%v at %v\n", t, d)
	}

	// print unmatched events
	if len(unmatched) > 0 {
		fmt.Printf("\nUnmatched events:\n")
		for _, u := range unmatched {
			fmt.Printf("  %v\n", u)
		}
	}

	// summarize billable usage
	billableUsage := float64(allocations["Billable"])/float64(totalTime)
	fmt.Printf("\nBillable usage is at %.0f%%\n", billableUsage * 100)

	// instruct how much billable usage is missing
	if billableUsage < 70 {
		moreBillableHours := time.Duration((0.70 - billableUsage) * float64(totalTime))
		fmt.Printf("\n%.5v more of your hours should be billable\n", moreBillableHours)
	}

	timeLeft := 40 * time.Hour - totalTime

	fmt.Printf("%v left in the week\n", timeLeft)
}
