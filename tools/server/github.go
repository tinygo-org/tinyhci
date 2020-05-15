package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v31/github"
)

func authenticateGithubClient(appid, installid int64, privatekeyfile string) (*github.Client, error) {
	tr := http.DefaultTransport
	itr, err := ghinstallation.NewKeyFromFile(tr, appid, installid, "keys/"+privatekeyfile)
	if err != nil {
		return nil, err
	}

	return github.NewClient(&http.Client{Transport: itr}), nil
}

func pendingCheckRun(sha string) {
	log.Printf("Github check run pending for %s\n", sha)
	opts := github.CreateCheckRunOptions{
		Name:    "TinyGo HCI",
		HeadSHA: sha,
	}
	cr, _, err := client.Checks.CreateCheckRun(context.Background(), ghorg, ghrepo, opts)
	if err != nil {
		log.Println(err)
	}
	runs[sha] = cr
}

func startCheckRun(sha string) {
	log.Printf("Github check run starting for %s\n", sha)
	status := "in_progress"
	if run, ok := runs[sha]; ok {
		opts := github.UpdateCheckRunOptions{
			Name:   "TinyGo HCI",
			Status: &status,
		}
		cr, _, err := client.Checks.UpdateCheckRun(context.Background(), ghorg, ghrepo, *run.ID, opts)
		if err != nil {
			log.Println(err)
		}
		runs[sha] = cr
	}
}

func passCheckRun(sha, output string) {
	log.Printf("Github check run pass for commit %s\n", sha)
	title := "TinyGo HCI"
	summary := "Hardware CI tests have failed."
	status := "completed"
	conclusion := "success"
	timestamp := github.Timestamp{Time: time.Now()}
	if run, ok := runs[sha]; ok {
		ro := github.CheckRunOutput{
			Title:   &title,
			Summary: &summary,
			Text:    &output,
		}

		opts := github.UpdateCheckRunOptions{
			Name:        "TinyGo HCI",
			Status:      &status,
			Conclusion:  &conclusion,
			CompletedAt: &timestamp,
			Output:      &ro,
		}
		_, _, err := client.Checks.UpdateCheckRun(context.Background(), ghorg, ghrepo, *run.ID, opts)
		if err != nil {
			log.Println(err)
		}
		delete(runs, sha)
	}
}

func failCheckRun(sha, output string) {
	log.Printf("Github check run fail for commit %s\n", sha)
	title := "TinyGo HCI"
	summary := "Hardware CI tests have failed."
	status := "completed"
	conclusion := "failure"
	timestamp := github.Timestamp{Time: time.Now()}
	if run, ok := runs[sha]; ok {
		ro := github.CheckRunOutput{
			Title:   &title,
			Summary: &summary,
			Text:    &output,
		}

		opts := github.UpdateCheckRunOptions{
			Name:        "TinyGo HCI",
			Status:      &status,
			Conclusion:  &conclusion,
			CompletedAt: &timestamp,
			Output:      &ro,
		}
		_, _, err := client.Checks.UpdateCheckRun(context.Background(), ghorg, ghrepo, *run.ID, opts)
		if err != nil {
			log.Println(err)
		}
		delete(runs, sha)
	}
}
