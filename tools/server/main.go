package main

import (
	"log"
	"os"
	"os/exec"

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
		payload, err := hook.Parse(r, github.PullRequestEvent)
		if err != nil {
			if err == github.ErrEventNotFound {
				// ok event wasn't one of the ones asked to be parsed
				log.Println("Not the event you are looking for")
				return
			}
		}
		switch payload.(type) {
		case github.PullRequestPayload:
			pullRequest := payload.(github.PullRequestPayload)
			log.Printf("%+v\n", pullRequest)
			out, err := exec.Command("make test-itsybitsy-m4").Output()
			if err != nil {
				log.Println(err)
				return
			}
			log.Printf("The output is %s\n", out)
		}
	})

	log.Println("Starting TinyHCI server...")
	http.ListenAndServe(":8000", nil)
}
