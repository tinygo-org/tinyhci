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

const (
	useCurrentBinaryRelease = false // set to true to use the already installed tinygo
	officialRelease         = "https://github.com/tinygo-org/tinygo/releases/download/v0.13.1/tinygo0.13.1.linux-amd64.tar.gz"
)

var (
	// these will be overwritten by the ENV vars of the same name
	ghorg  = "tinygo-org"
	ghrepo = "tinygo"

	ghwebhookpath = "/webhooks"
	ciwebhookpath = "/buildhook"

	client *github.Client

	// key is sha
	builds map[string]*Build

	pollFrequency = 30 * time.Second
)

func main() {
	ghwebhookpath = os.Getenv("GHWEBHOOKPATH")
	if ghwebhookpath == "" {
		log.Fatal("You must set an ENV var with your GHWEBHOOKPATH")
	}

	ciwebhookpath = os.Getenv("CIWEBHOOKPATH")
	if ciwebhookpath == "" {
		log.Fatal("You must set an ENV var with your CIWEBHOOKPATH")
	}

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

	builds = make(map[string]*Build)
	buildsCh := make(chan *Build)

	go processBuilds(buildsCh)
	go pollPendingBuilds(buildsCh)

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
			// ignore pushes because we care about checks
			return
		case *github.CheckSuiteEvent:
			log.Printf("Github checksuite event %s %s for %d %s\n",
				event.CheckSuite.GetStatus(),
				event.CheckSuite.GetConclusion(),
				event.CheckSuite.GetID(),
				event.CheckSuite.GetHeadSHA())

			// ignore completed events to avoid endless loop
			if event.CheckSuite.GetStatus() == "completed" {
				return
			}

			// received when a new commit is pushed
			build := NewBuild(event.CheckSuite.GetHeadSHA())
			build.pendingCI = true
			build.started = time.Now()
			builds[build.sha] = build
			build.pendingCheckSuite()

		case *github.CheckRunEvent:
			log.Printf("Github checkrun event %s %s for %d %s %s\n",
				event.CheckRun.GetStatus(),
				event.CheckRun.GetConclusion(),
				event.CheckRun.GetID(),
				event.CheckRun.GetName(),
				event.CheckRun.GetHeadSHA())

			// ignore completed events to avoid endless loop
			if event.CheckRun.GetStatus() == "completed" {
				return
			}

			// received when we are asked to re-run a failed check run
			var build *Build
			target, err := parseTarget(event.CheckRun.GetName())
			if err != nil {
				log.Println(err)
				return
			}
			board := GetBoard(target)

			// first check to see if this build is in cache
			if build, ok := builds[event.CheckRun.GetHeadSHA()]; ok {
				build.processBoardRun(board)
				return
			}

			// if not, then create new build
			bn, err := getCIBuildNumFromSHA(event.CheckRun.GetHeadSHA())
			if err != nil {
				log.Println(err)
				return
			}

			url, err := getTinygoBinaryURL(bn)
			if err != nil {
				log.Println(err)
				return
			}

			build = NewBuild(event.CheckRun.GetHeadSHA())
			build.binaryURL = url
			build.runs[target] = event.CheckRun
			builds[build.sha] = build

			// handoff to channel for processing
			buildsCh <- build

		default:
			log.Println("Unexpected Github event:", event)
		}
	})

	// we can remove this soon.
	http.HandleFunc(ciwebhookpath, func(w http.ResponseWriter, r *http.Request) {
		log.Println("CircleCI buildhook received.")
		bi, err := parseBuildInfo(r)
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("Build Info: %+v\n", bi)
	})

	log.Printf("Starting TinyHCI server for %s/%s\n", ghorg, ghrepo)
	http.ListenAndServe(":8000", nil)
}

// processBuilds is run as a go routine to pull new builds
// from the build channel, and then perform the needed build
// tasks aka build docker image, then flash/test for each board.
func processBuilds(builds chan *Build) {
	for {
		select {
		case build := <-builds:
			log.Printf("Starting tests for commit %s\n", build.sha)
			build.startCheckSuite()

			url := officialRelease
			if !useCurrentBinaryRelease {
				url = build.binaryURL
			}

			log.Printf("Building docker image using TinyGo from %s\n", url)
			err := buildDocker(url, build.sha)
			if err != nil {
				log.Println(err)
				build.failCheckSuite("docker build failed")
				continue
			}

			log.Printf("Running checks for commit %s\n", build.sha)
			for _, run := range build.runs {
				target, _ := parseTarget(run.GetName())
				board := GetBoard(target)
				build.processBoardRun(board)
			}
		}
	}
}

// buildDocker does the docker build for the binary download
// with this SHA.
func buildDocker(url, sha string) error {
	buildarg := fmt.Sprintf("TINYGO_DOWNLOAD_URL=%s", url)
	buildtag := "tinygohci:" + sha[:7]
	out, err := exec.Command("docker", "build",
		"-t", buildtag,
		"-f", "tools/docker/Dockerfile",
		"--build-arg", buildarg, ".").CombinedOutput()
	if err != nil {
		log.Println(err)
		log.Println(string(out))
		return err
	}

	return nil
}

// pollPendingBuilds is run as a go routine to poll the
// circleci server, and look for new builds of the TinyGo
// binary that match tinyhci builds in need of processing.
func pollPendingBuilds(buildsCh chan *Build) {
	for {
		if len(builds) == 0 {
			log.Println("No builds to poll for.")
		} else {
			log.Println("Polling for builds...")
		}

		for _, b := range builds {
			if b.pendingCI {
				// look to see if there is a CI build with binary for this build
				bn, err := getMostRecentCIBuildNumAfterStart(b.sha, b.started)
				if err != nil {
					continue
				}
				log.Println("Binary ready for", b.sha)

				url, err := getTinygoBinaryURL(bn)
				if err != nil {
					log.Println(err)
					continue
				}

				b.binaryURL = url
				b.pendingCI = false

				buildsCh <- b
			}
		}

		// sleep in-between waiting for new builds
		time.Sleep(pollFrequency)
	}
}
