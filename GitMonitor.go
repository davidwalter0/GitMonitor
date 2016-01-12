package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

/*
TODO
Testing?
*/

const minsToWait int = 1 // Number of minutes to wait between each update

// A structure to keep track of relevant data
type GitHubEvent struct {
	Username string
	Type     string
	Repo     string
	EventID  string
	Public   bool
	Day      int
	Month    time.Month
	Year     int
}

// Overload the String() function to present gitHubEvents in an easy-to-parse manner
func (g GitHubEvent) String() string {
	return fmt.Sprintf("- %v commmited a %v on the repo %q on %v %v, %v with ID %q (public: %v)", g.Username, g.Type, g.Repo, g.Month, g.Day, g.Year, g.EventID, g.Public)
}

// Returns a GitHubEvent struct from the go-github github.Event struct
func getGitHubEvent(cur github.Event) GitHubEvent {
	return GitHubEvent{*cur.Actor.Login, *cur.Type, *cur.Repo.Name, *cur.ID, *cur.Public, cur.CreatedAt.Day(), cur.CreatedAt.Month(), cur.CreatedAt.Year()}
}

// Listen for events either by owner or repository and prints new activity to standard out, updates every minsToWait mins
func ListenForEvents(owner, repo string, ch chan<- GitHubEvent, errorChannel chan<- error) {
	// Set up the go-github client and authenticate
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ""}) // Insert your personal access token here
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)
	_, gitResponse, err := client.APIMeta()
	if err != nil {
		errorChannel <- err
		return
	}
	// Check for status
	if gitResponse.Remaining == 0 {
		errorChannel <- errors.New("No API calls remaining")
		return
	}
	// Represents the most recent previously seen event
	var previousEvent GitHubEvent
	for {
		// Default to monitoring user events & check to ensure that the user has entered both the username and/or repository name correctly
		// Move out to a function
		events, _, err := client.Activity.ListEventsPerformedByUser(owner, false, nil)
		if err != nil {
			errorChannel <- err
			break
		}
		if repo != "" {
			events, _, err = client.Activity.ListRepositoryEvents(owner, repo, nil)
			if err != nil {
				errorChannel <- err
				break
			}
		}
		first := getGitHubEvent(events[0])
		if first != previousEvent {
			mostRecentlySeen := first
			toSendToChan := make([]GitHubEvent, 0)
			for _, val := range events {
				cur := getGitHubEvent(val)
				if cur == previousEvent {
					break
				}
				// If it's new, add it to the slice
				toSendToChan = append(toSendToChan, cur)
			}
			// todo: could potentially rephrase syntax to use a more traditional for loop
			for i := range toSendToChan {
				// Sends items to channel in reverse order so new events are printed chronologically
				ch <- toSendToChan[len(toSendToChan)-i-1]
			}
			previousEvent = mostRecentlySeen
		}

		time.Sleep(time.Minute * time.Duration(minsToWait))
	}
}

// Prints every GitHubEvent received through the channel
func PrintEvents(ch <-chan GitHubEvent, userActivityMap map[string]int) {
	for {
		received := <-ch
		userActivityMap[received.Username]++
		fmt.Println(received)
	}
}

// Listens for errors sent through the channel signaling that something went wrong with the API call
func ListenForErrors(ch <-chan error) {
	for {
		receivedError := <-ch
		fmt.Printf("An error occurred, it was: %q\n", receivedError)
	}
}

func main() {
	toPrint := make(chan GitHubEvent)
	catchErrors := make(chan error)
	activityMap := make(map[string]int) // Some prefer curly brace initialization
	go PrintEvents(toPrint, activityMap)
	go ListenForErrors(catchErrors)
loop:
	for {
		fmt.Println("Would you like to monitor a user or a repo? Please enter 1 for user, 2 for repo or 3 to exit")
		var input int
		fmt.Scanln(&input)
		switch input {
		case 1:
			var user string
			fmt.Println("Please enter their github username")
			fmt.Scanln(&user)
			go ListenForEvents(user, "", toPrint, catchErrors)
		case 2:
			var owner string
			fmt.Println("Please enter the github username of the owner of the repo")
			fmt.Scanln(&owner)
			var repoName string
			fmt.Println("Please enter the name of the repo")
			fmt.Scanln(&repoName)
			go ListenForEvents(owner, repoName, toPrint, catchErrors)
		case 3:
			break loop
		default:
			fmt.Println("You entered something else, please try again")

		}
	}
	if len(activityMap) != 0 {
		fmt.Println("The commit/event count for each user is as follows:\n", activityMap)
	}
	fmt.Println("Exiting")
}
