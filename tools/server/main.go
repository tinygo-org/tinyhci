package main

import (
	"fmt"
	"os"

	"net/http"

	"gopkg.in/go-playground/webhooks.v5/github"
)

const (
	path = "/webhooks"
)

func main() {
	ghkey := os.Getenv("GHKEY")
	if ghkey == "" {
		panic("You must set an ENV var with your GHKEY")
	}

	hook, _ := github.New(github.Options.Secret(ghkey))

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, github.ReleaseEvent, github.PullRequestEvent)
		if err != nil {
			if err == github.ErrEventNotFound {
				// ok event wasn't one of the ones asked to be parsed
				fmt.Println("Not the event you are looking for")
			}
		}
		switch payload.(type) {
		case github.ReleasePayload:
			release := payload.(github.ReleasePayload)
			// Do whatever you want from here...
			fmt.Printf("%+v", release)

		case github.PullRequestPayload:
			pullRequest := payload.(github.PullRequestPayload)
			// Do whatever you want from here...
			fmt.Printf("%+v", pullRequest)
		}
	})
	http.ListenAndServe(":8000", nil)
}
