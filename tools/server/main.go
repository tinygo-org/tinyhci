package main

import (
	"log"
	"os"
	"os/exec"

	"net/http"

	"gopkg.in/go-playground/webhooks.v5/github"
)

const (
	path    = "/webhooks"
	testCmd = "make test-itsybitsy-m4"
)

func main() {
	ghkey := os.Getenv("GHKEY")
	if ghkey == "" {
		panic("You must set an ENV var with your GHKEY")
	}

	builds := make(chan string)

	go processBuilds(builds)

	hook, _ := github.New(github.Options.Secret(ghkey))
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, github.PushEvent)
		if err != nil {
			if err == github.ErrEventNotFound {
				// ok event wasn't one of the ones asked to be parsed
				log.Println("Not the event you are looking for")
				return
			}
		}
		switch payload.(type) {
		case github.PushPayload:
			push := payload.(github.PushPayload)
			builds <- push.After
		}
	})

	log.Println("Starting TinyHCI server...")
	http.ListenAndServe(":8000", nil)
}

func processBuilds(builds chan string) {
	for {
		select {
		case build := <-builds:
			log.Printf("Running tests for commit %s\n", build)
			out, err := exec.Command("sh", "-c", testCmd).CombinedOutput()
			if err != nil {
				log.Println(err)
				log.Println(string(out))
				continue
			}
			log.Printf(string(out))
		}
	}
}
