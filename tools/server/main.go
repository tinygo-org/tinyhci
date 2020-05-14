package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"

	"net/http"

	"github.com/google/go-github/v31/github"
)

const (
	path    = "/webhooks"
	testCmd = "make test-itsybitsy-m4"
)

type Build struct {
	binaryUrl string
	sha       string
}

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

	client, err = authenticateGithubClient(int64(appid), int64(installid), ghkeyfile)
	if err != nil {
		log.Println(err)
	}

	builds := make(chan *Build)
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
			pendingCheckRun(*event.After)
		default:
			log.Println("Unexpected Github event")
		}
	})

	http.HandleFunc("/buildhook", func(w http.ResponseWriter, r *http.Request) {
		log.Println("CircleCI buildhook received.")
		bi, err := parseBuildInfo(r)
		if err != nil {
			log.Println(err)
			return
		}

		log.Printf("Build Info: %+v\n", bi)
		url, err := getTinygoBinaryURL(bi.BuildNum)
		if err != nil {
			log.Println(err)
			return
		}

		builds <- &Build{sha: bi.VcsRevision, binaryUrl: url}
	})

	log.Println("Starting TinyHCI server...")
	http.ListenAndServe(":8000", nil)
}

func processBuilds(builds chan *Build) {
	for {
		select {
		case build := <-builds:
			log.Printf("Starting tests for commit %s\n", build.sha)
			startCheckRun(build.sha)

			log.Printf("Downloading new TinyGo from %s\n", build.binaryUrl)
			downloadFile("/tmp/tinygo.tar.gz", build.binaryUrl)

			log.Printf("Installing TinyGo from commit %s\n", build.sha)
			installBinary("/tmp/tinygo.tar.gz")

			log.Printf("Running tests for commit %s\n", build.sha)
			out, err := exec.Command("sh", "-c", testCmd).CombinedOutput()
			if err != nil {
				log.Println(err)
				log.Println(string(out))
				failCheckRun(build.sha)
				continue
			}
			passCheckRun(build.sha)
			log.Printf(string(out))
		}
	}
}

func downloadFile(filepath string, url string) (err error) {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func installBinary(filename string) error {
	out, err := exec.Command("tar", "-xzf", filename, "-C", "/usr/local").CombinedOutput()
	if err != nil {
		log.Println(err)
		log.Println(string(out))
		return err
	}

	return nil
}
