package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

/*
TODO
Testing?
Find a way to get commit message? -> DONE
	Need to figure out why it crashes/panics after any more than 3 iterations
Find a way to read the list of events so they print in chronological order -> DONE
Abstract --> DONE


*/

const minsToWait int = 5 // Number of minutes to wait between each update

// A structure to keep track of relevant data
type GitHubEvent struct {
	Name     string
	Username string
	Type     string
	Repo     string
	EventID  string
	Public   bool
	Message  string
}

// Overload the String() function to present the data in an easy-to-parse manner
func (g GitHubEvent) String() string {
	return fmt.Sprintf("- %v (%v) commmited a %v on the repo %q with message %q and ID %q (public: %v)", g.Name, g.Username, g.Type, g.Repo, g.Message, g.EventID, g.Public)
}

// Returns a GitHubEvent struct from the go-github github.Event struct
func getGitHubEvent(cur github.Event) GitHubEvent {
	var data map[string]interface{}                                // Declare a map to parse the json
	if err := json.Unmarshal(*cur.RawPayload, &data); err != nil { // Unmarshal the JSON
		panic(err)
	}
	commitArray := data["commits"].([]interface{})       // Cast the "commit" entry of the map to an array
	commitMap := commitArray[0].(map[string]interface{}) // Grab the first element of the array, cast it to a map[string]interface{} representing the commit information
	/*
		Other fields in commitmap besides "message" include: "distinct", url", "sha", "author" (returns a map with fields "email" and "name") )
	*/
	commitMessage := commitMap["message"].(string) // Typecast interface{} to string
	authorMap := commitMap["author"].(map[string]interface{})
	authorName := authorMap["name"].(string) // Typecast interface{} to string
	return GitHubEvent{authorName, *cur.Actor.Login, *cur.Type, *cur.Repo.Name, *cur.ID, *cur.Public, commitMessage}
}

// Listen for events on a particular repo, updates every minsToWait mins
func ListenForEvents(owner, repo string, ch chan<- GitHubEvent) {
	ts := oauth2.StaticTokenSource( // Set up client
		&oauth2.Token{AccessToken: "2f39da9a4134b244b909fa38527005a2c953b9c7"}) // Insert your personal access token here
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)
	var prev GitHubEvent
	for {
		events, _, _ := client.Activity.ListEventsPerformedByUser(owner, false, nil) // Default to monitoring user events
		if repo != "" {                                                              // If repo is not nil, monitory repo events instead
			events, _, _ = client.Activity.ListRepositoryEvents(owner, repo, nil)
		}
		cur := getGitHubEvent(events[0])
		if cur != prev {
			mostRec := cur
			toSendToChan := make([]GitHubEvent, 0) // Add all events to a slice
			for i := 0; i < 3; i++ {
				rn := getGitHubEvent(events[i])
				if rn == prev {
					break
				}
				toSendToChan = append(toSendToChan, rn)
			}
			for i, _ := range toSendToChan {
				ch <- toSendToChan[len(toSendToChan)-i-1] // Sends items to channel in reverse order so new events appear chronologically at the bottom
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
	}

	fmt.Println("Exiting")
}
