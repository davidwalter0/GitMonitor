package main

import (
	"fmt"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

/*
TODO
Testing?
Find a way to get commit message?
Find a way to read the list of events so they print in chronological order
Abstract --> DONE


*/

const minsToWait int = 1 // Number of minutes to wait between each update

// A structure to keep track of relevant data
type GitHubEvent struct {
	Username string
	Type     string
	Repo     string
	EventID  string
}

// Overload the String() function to present the data in an easy-to-parse manner
func (g GitHubEvent) String() string {
	return fmt.Sprintf("%v commmited a %v on the repo %v with ID %v", g.Username, g.Type, g.Repo, g.EventID)
}

// Returns a GitHubEvent struct from the go-github github.Event struct
func getGitHubEvent(cur github.Event) GitHubEvent {
	return GitHubEvent{*cur.Actor.Login, *cur.Type, *cur.Repo.Name, *cur.ID}
}

// Listen for events on a particular repo, updates every minsToWait mins
func ListenForEvents(owner, repo string, ch chan<- GitHubEvent) {
	ts := oauth2.StaticTokenSource( // Set up client
		&oauth2.Token{AccessToken: ""}) // Insert your personal access token here
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)
	var prev GitHubEvent
	for {
		events, _, _ := client.Activity.ListEventsPerformedByUser(owner, false, nil)
		if repo != "" {
			//	events, _, _ := client.Activity.ListRepositoryEvents(owner, repo, nil)
		}
		cur := getGitHubEvent(events[0])
		fmt.Println(cur)
		if cur != prev {
			mostRec := cur
			for _, v := range events {
				rn := getGitHubEvent(v)
				if rn == prev {
					break
				}
				ch <- rn
			}
			prev = mostRec // Update prev to the most recently encountered item on the list

		}
		time.Sleep(time.Minute * time.Duration(minsToWait)) // Sleep for minsToWait mins
	}
}

// Prints every GitHubEvent received through the channel
func PrintEvents(ch <-chan GitHubEvent) {
	for {
		cur := <-ch
		fmt.Println(cur)
	}
}

func main() {
	toPrint := make(chan GitHubEvent)
	go PrintEvents(toPrint)
	go ListenForEvents("ElTav", "", toPrint)
	/*
		for {
			fmt.Println("Would you like to monitor a user or a repo? Please enter 1 for user, 2 for repo or 3 to exit")
			var input int
			fmt.Scanln(&input)
			if input == 1 {
				var user string
				fmt.Println("Please enter their github username")
				fmt.Scanln(&user)
				go ListenForEvents(user, "", toPrint)
			} else if input == 2 {
				var owner string
				fmt.Println("Please enter the github username of the owner of the repo")
				fmt.Scanln(&owner)
				var repoName string
				fmt.Println("Please enter the name of the repo")
				fmt.Scanln(&repoName)
				go ListenForEvents(owner, repoName, toPrint)
			} else if input == 3 {
				break
			} else {
				fmt.Println("You entered something else, please try again")
			}
		}*/
	var user string
	fmt.Println("Please enter their github username")
	fmt.Scanln(&user)
	fmt.Println("Exiting")
}
