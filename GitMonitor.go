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
Find a way to get commit message? -> DONE
	Need to figure out why it crashes/panics after any more than 3 iterations -> DONE
	Note to self: Can access commit data and IFF it's for your account and repo, else not allowed + crashes
Find a way to read the list of events so they print in chronological order -> DONE
Abstract --> DONE


*/

const minsToWait int = 1 // Number of minutes to wait between each update

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
	return fmt.Sprintf("- %v commmited a %v on the repo %q and ID %q (public: %v)", g.Username, g.Type, g.Repo, g.EventID, g.Public)
}

// Returns a GitHubEvent struct from the go-github github.Event struct
func getGitHubEvent(cur github.Event) GitHubEvent {
	/*v Note: For below data, see note in todo section
	ar data map[string]interface{}                                // Declare a map to parse the json
	if err := json.Unmarshal(*cur.RawPayload, &data); err != nil { // Unmarshal the JSON
		panic(err)
	}
	commitArray := data["commits"].([]interface{})       // Cast the "commit" entry of the map to an array
	commitMap := commitArray[0].(map[string]interface{}) // Grab the first element of the array, cast it to a map[string]interface{} representing the commit information
	fmt.Println(commitMap)
	/*
		Other fields in commitmap besides "message" include: "distinct", url", "sha", "author" (returns a map with fields "email" and "name") )
	*/
	//commitMessage := commitMap["message"].(string) // Typecast interface{} to string
	//fmt.Println(commitMessage)
	//authorMap := commitMap["author"].(map[string]interface{})
	//fmt.Println(authorMap)
	//authorName := authorMap["name"].(string) // Typecast interface{} to string
	return GitHubEvent{"authorName", *cur.Actor.Login, *cur.Type, *cur.Repo.Name, *cur.ID, *cur.Public, "commitMessage"}
}

// Listen for events on a particular repo, updates every minsToWait mins
func ListenForEvents(owner, repo string, ch chan<- GitHubEvent) {
	ts := oauth2.StaticTokenSource( // Set up client
		&oauth2.Token{AccessToken: ""}) // Insert your personal access token here
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)
	var prev GitHubEvent
	for {
		events, _, _ := client.Activity.ListEventsPerformedByUser(owner, false, nil) // Default to monitoring user events
		if repo != "" {                                                              // If repo is not nil, monitory repo events instead
			events, _, _ = client.Activity.ListRepositoryEvents(owner, repo, nil)
		}

		first := getGitHubEvent(events[0])
		if first != prev {
			mostRec := first
			toSendToChan := make([]GitHubEvent, 0) // Add all events to a slice
			for _, val := range events {
				cur := getGitHubEvent(val)
				if cur == prev {
					break
				}
				toSendToChan = append(toSendToChan, cur)
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
