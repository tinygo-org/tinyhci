package main

import (
	"log"
	"os"
	"os/exec"
	"strconv"

	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v31/github"
)

const (
	path    = "/webhooks"
	testCmd = "make test-itsybitsy-m4"
)

var (
	client *github.Client
)

func main() {
	ghkey := os.Getenv("GHKEY")
	if ghkey == "" {
		log.Fatal("You must set an ENV var with your GHKEY")
	}

	ghkeyfile := os.Getenv("GHKEYFILE")
	if ghkeyfile == "" {
		log.Fatal("You must set an ENV var with your GHKEYFILE")
	}

	aid := os.Getenv("GHAPPID")
	if aid == "" {
		log.Fatal("You must set an ENV var with your GHAPPID")
	}

	iid := os.Getenv("GHINSTALLID")
	if iid == "" {
		log.Fatal("You must set an ENV var with your GHINSTALLID")
	}

	appid, err := strconv.Atoi(aid)
	if err != nil {
		log.Fatal("Invalid Github app id")
	}

	installid, err := strconv.Atoi(iid)
	if err != nil {
		log.Fatal("Invalid Github install id")
	}

	client, err = authenticateClient(int64(appid), int64(installid), ghkeyfile)
	if err != nil {
		log.Println(err)
	}

	builds := make(chan string)
	go processBuilds(builds)

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		payload, err := github.ValidatePayload(r, []byte(ghkey))
		if err != nil {
			log.Println("Invalid webhook payload")
			return
		}
		event, err := github.ParseWebHook(github.WebHookType(r), payload)
		if err != nil {
			log.Println("Invalid webhook event")
			return
		}
		switch event := event.(type) {
		case *github.PushEvent:
			builds <- *event.After
		default:
			log.Println("Not the event you are looking for")
		}
	})

	http.HandleFunc("/buildhook", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got the buildhook:")
		bi, err := parseBuildInfo(r)
		if err != nil {
			log.Println(err)
		}

		log.Printf("Build Info: %+v\n", bi)
		getTinygoBinary(bi.BuildNum)
		return
	})

	log.Println("Starting TinyHCI server...")
	http.ListenAndServe(":8000", nil)
}

func processBuilds(builds chan string) {
	for {
		select {
		case build := <-builds:
			log.Printf("Running tests for commit %s\n", build)
			startCheckRun()

			out, err := exec.Command("sh", "-c", testCmd).CombinedOutput()
			if err != nil {
				log.Println(err)
				log.Println(string(out))
				failCheckRun()
				continue
			}
			passCheckRun()
			log.Printf(string(out))
		}
	}
}

func authenticateClient(appid, installid int64, privatekeyfile string) (*github.Client, error) {
	tr := http.DefaultTransport
	itr, err := ghinstallation.NewKeyFromFile(tr, appid, installid, "keys/"+privatekeyfile)
	if err != nil {
		return nil, err
	}

	return github.NewClient(&http.Client{Transport: itr}), nil
}

func startCheckRun() {
}

func passCheckRun() {
}

func failCheckRun() {
}
