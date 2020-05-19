package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

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
		case *github.CheckSuiteEvent:
			// received when a new commit is pushed
			build := NewBuild(event.CheckSuite.GetHeadSHA())
			builds[build.sha] = build
			build.pendingCheckSuite()

		case *github.CheckRunEvent:
			log.Printf("Github checkrun event %s for %d %s, %s", event.CheckRun.GetStatus(), event.CheckRun.GetID(), event.CheckRun.GetName(), event.CheckRun.GetHeadSHA())

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

		// first check to see if this build is in cache
		var build *Build
		ok := false
		if build, ok = builds[bi.VCSRevision]; !ok {
			// if not, then create new build
			build = NewBuild(bi.VCSRevision)
			builds[build.sha] = build
			// get the runs for this build
			err := build.reloadCheckRuns()
			if err != nil {
				log.Println(err)
				return
			}
		}

		build.binaryURL = url

		buildsCh <- build
	})

	log.Printf("Starting TinyHCI server for %s/%s\n", ghorg, ghrepo)
	http.ListenAndServe(":8000", nil)
}

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
