package main

import (
	"path/filepath"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"net/http"

	"github.com/cavaliercoder/grab"
	"github.com/google/go-github/v40/github"
)

const (
	useCurrentBinaryRelease = false // set to true to use the already installed tinygo
	officialRelease         = "https://github.com/tinygo-org/tinygo/releases/download/v0.21.0/tinygo0.21.0.linux-amd64.tar.gz"
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

	// start go routine to actually do the building
	go processBuilds(buildsCh)

	// fetch any builds that are already in progress
	go handlePreviouslyQueuedBuilds(buildsCh)

	// start the webhook server
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
			// ignore pushes because we only care about checks API
			return
		case *github.WorkflowRunEvent:
			log.Printf("Github workflowrun on '%s' event %s %s for %d %s\n",
				event.WorkflowRun.GetName(),
				event.WorkflowRun.GetStatus(),
				event.WorkflowRun.GetConclusion(),
				event.WorkflowRun.GetID(),
				event.WorkflowRun.GetHeadSHA())

			if event.WorkflowRun.GetStatus() == "completed" &&
				event.WorkflowRun.GetConclusion() == "success" &&
				event.WorkflowRun.GetName() == "Linux" {
				url, err := getTinygoBinaryURLFromGH(event.WorkflowRun.GetID())
				if err != nil {
					log.Println(err)
					return
				}

				b, ok := builds[event.WorkflowRun.GetHeadSHA()]
				if !ok {
					b = NewBuild(event.WorkflowRun.GetHeadSHA())
					builds[event.WorkflowRun.GetHeadSHA()] = b
				}
				b.binaryURL = url
				b.pendingCI = false
				buildsCh <- b
			}

		case *github.WorkflowJobEvent:
			log.Printf("Github workflowjob on '%s' event %s %s for %d %s\n",
				event.WorkflowJob.GetName(),
				event.WorkflowJob.GetStatus(),
				event.WorkflowJob.GetConclusion(),
				event.WorkflowJob.GetID(),
				event.WorkflowJob.GetHeadSHA())

		case *github.CheckSuiteEvent:
			log.Printf("Github checksuite event %s %s for %d %s\n",
				event.CheckSuite.GetStatus(),
				event.CheckSuite.GetConclusion(),
				event.CheckSuite.GetID(),
				event.CheckSuite.GetHeadSHA())

			switch event.CheckSuite.GetStatus() {
			case "completed":
				// just in case we want to do something here
			case "queued":
				// received when a new commit is pushed
				build := NewBuild(event.CheckSuite.GetHeadSHA())
				build.pendingCI = true
				build.started = time.Now()
				builds[build.sha] = build
				build.pendingCheckSuite()
			default:
			}

		case *github.CheckRunEvent:
			log.Printf("Github checkrun event %s %s for %d %s %s %s %s %s\n",
				event.CheckRun.GetStatus(),
				event.CheckRun.GetConclusion(),
				event.CheckRun.GetID(),
				event.CheckRun.GetName(),
				event.GetAction(),
				event.CheckRun.GetHeadSHA(),
				event.CheckRun.GetExternalID(),
				event.CheckRun.GetDetailsURL())

			switch event.CheckRun.GetStatus() {
			case "completed":
				if event.GetAction() == "rerequested" {
					wr, err := getRecentWorkflowRunForSHA("success", event.CheckRun.GetHeadSHA())
					if err != nil {
						log.Println(err)
						return
					}

					performCheckRun(event.CheckRun, wr.GetID(), buildsCh)
				}

			case "queued":
				// received when a new commit is pushed
			default:
			}

		default:
			log.Println("Unexpected Github event:", event)
		}
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

			log.Printf("Downloading TinyGo from %s\n", url)
			err := downloadBinary(url, build.sha)
			if err != nil {
				log.Println(err)
				build.failCheckSuite("binary download failed")
				continue
			}

			log.Printf("Building docker image using TinyGo from %s\n", url)
			err = buildDocker(build.sha)
			if err != nil {
				log.Println(err)
				build.failCheckSuite("docker build failed")
				continue
			}

			log.Printf("Running checks for commit %s\n", build.sha)
			for _, run := range build.runs {
				target, err := parseTarget(run.GetName())
				if err != nil {
					log.Println(err)
					build.failCheckRun(target, err.Error())
					continue
				}
				board := GetBoard(target)
				if board != nil {
					build.processBoardRun(board)
				}
			}
		}
	}
}

// buildDocker does the docker build for the binary download
// with this SHA.
func buildDocker(sha string) error {
	buildarg := fmt.Sprintf("TINYGO_DOWNLOAD_SHA=%s", sha)
	buildtag := "tinygohci:" + sha[:7]
	cmd := exec.Command("docker", "build",
		"-t", buildtag,
		"-f", "tools/docker/Dockerfile",
		"--build-arg", buildarg, ".")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "DOCKER_BUILDKIT=1")

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
		log.Println(string(out))
		return err
	}

	return nil
}

// downloadBinary does the download for the binary build
// with this SHA.
func downloadBinary(url, sha string) error {
	// check if the file is already downloaded for this sha
	if !fileExists("tools/docker/versions/" + sha + ".tar.gz") {
		log.Println("Downloading binary for", sha)

		resp, err := grab.Get("tinygo-latest.zip", url)
		if err != nil {
			return err
		}
		log.Println("downloaded bytes:", resp.BytesComplete())
		// unzip
		log.Println("unzipping")
		out, err := exec.Command("unzip", "tinygo-latest.zip",
			"tinygo*.linux-amd64.tar.gz").CombinedOutput()
		if err != nil {
			return err
		}
		log.Println(string(out))

		// move file
		log.Println("moving file")
		f, err := filepath.Glob("tinygo*.linux-amd64.tar.gz")
		if err != nil {
			return err
		}
		err = os.Rename(f[0], "tools/docker/versions/"+sha+".tar.gz")
		if err != nil {
			return err
		}

		err = os.Remove("tinygo-latest.zip")
		if err != nil {
			return err
		}
	}
	return nil
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func performCheckRun(cr *github.CheckRun, runID int64, buildsCh chan *Build) {
	// do the retest here
	url, err := getTinygoBinaryURLFromGH(runID)
	if err != nil {
		log.Println(err)
		return
	}

	target, err := parseTarget(cr.GetName())
	if err != nil {
		log.Println(err)
		return
	}

	build := NewBuild(cr.GetHeadSHA())
	build.binaryURL = url
	build.runs[target] = cr
	builds[build.sha] = build

	// handoff to channel for processing
	buildsCh <- build
}

// handlePreviouslyQueuedBuilds retrieves builds that were
// already queued before the server was started, probably due
// to some error or failure.
func handlePreviouslyQueuedBuilds(buildsCh chan *Build) {
	cibuilds, err := getRecentSuccessfulWorkflowRuns()
	if err != nil {
		log.Println(err)
		return
	}

	for _, cib := range cibuilds {
		// any in_progress checkruns for this build? restart them
		runs, err := findCheckRuns(cib.GetHeadSHA(), "in_progress")
		if err != nil {
			log.Println(err)
			return
		}

		for _, run := range runs {
			performCheckRun(run, cib.GetID(), buildsCh)
		}
	}

	for _, cib := range cibuilds {
		// any queued checkruns for this build?
		runs, err := findCheckRuns(cib.GetHeadSHA(), "queued")
		if err != nil {
			log.Println(err)
			return
		}

		for _, run := range runs {
			performCheckRun(run, cib.GetID(), buildsCh)
		}
	}
}
