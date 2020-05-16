package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"net/http"

	"github.com/google/go-github/v31/github"
)

// Build is a specific build to be tested.
type Build struct {
	binaryURL string
	sha       string
}

const (
	debugSkipBinaryInstall = true // set to true to use the already installed tinygo
	officialRelease        = "https://github.com/tinygo-org/tinygo/releases/download/v0.13.1/tinygo0.13.1.linux-amd64.tar.gz"
)

var (
	// these will be overwritten by the ENV vars of the same name
	ghorg  = "tinygo-org"
	ghrepo = "tinygo"

	ghwebhookpath = "/webhooks"
	ciwebhookpath = "/buildhook"
	testCmd       = "make test-itsybitsy-m4"

	client *github.Client
	runs   map[string]*github.CheckRun
)

func main() {
	ghorg = os.Getenv("GHORG")
	if ghorg == "" {
		log.Fatal("You must set an ENV var with your GHORG")
	}

	ghrepo = os.Getenv("GHREPO")
	if ghrepo == "" {
		log.Fatal("You must set an ENV var with your GHREPO")
	}

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

	client, err = authenticateGithubClient(int64(appid), int64(installid), ghkeyfile)
	if err != nil {
		log.Println(err)
	}

	runs = make(map[string]*github.CheckRun)

	builds := make(chan *Build)
	go processBuilds(builds)

	http.HandleFunc(ghwebhookpath, func(w http.ResponseWriter, r *http.Request) {
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
			pendingCheckRun(*event.After)
		default:
			log.Println("Unexpected Github event")
		}
	})

	http.HandleFunc(ciwebhookpath, func(w http.ResponseWriter, r *http.Request) {
		log.Println("CircleCI buildhook received.")
		bi, err := parseBuildInfo(r)
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("Build Info: %+v\n", bi)
		if bi.Status != "success" {
			log.Printf("Not running tests for %s status was %s\n", bi.VCSRevision, bi.Status)
			return
		}

		url, err := getTinygoBinaryURL(bi.BuildNum)
		if err != nil {
			log.Println(err)
			return
		}

		builds <- &Build{sha: bi.VCSRevision, binaryURL: url}
	})

	log.Printf("Starting TinyHCI server for %s/%s\n", ghorg, ghrepo)
	http.ListenAndServe(":8000", nil)
}

func processBuilds(builds chan *Build) {
	for {
		select {
		case build := <-builds:
			log.Printf("Starting tests for commit %s\n", build.sha)
			startCheckRun(build.sha)

			url := officialRelease
			if !debugSkipBinaryInstall {
				url = build.binaryURL
			}

			log.Printf("Building docker image using TinyGo from %s\n", url)
			err := buildDocker(url)
			if err != nil {
				log.Println(err)
				failCheckRun(build.sha, "docker build failed")
				continue
			}

			log.Printf("Running checks for commit %s\n", build.sha)
			for _, board := range boards {
				log.Printf("Flashing board %s\n", board.displayname)
				flashout, err := board.flash()
				if err != nil {
					log.Println(err)
					log.Println(flashout)
					failCheckRun(build.sha, flashout)
					continue
				}

				time.Sleep(2 * time.Second)

				log.Printf("Running tests on board %s\n", board.displayname)
				out, err := board.test()
				if err != nil {
					log.Println(err)
					log.Println(out)
					failCheckRun(build.sha, out)
					continue
				}
				passCheckRun(build.sha, out)
			}
		}
	}
}

func buildDocker(url string) error {
	buildarg := fmt.Sprintf("TINYGO_DOWNLOAD_URL=%s", url)
	out, err := exec.Command("docker", "build",
		"-t", "tinygohci",
		"-f", "tools/docker/Dockerfile",
		"--build-arg", buildarg, ".").CombinedOutput()
	if err != nil {
		log.Println(err)
		log.Println(string(out))
		return err
	}

	return nil
}
