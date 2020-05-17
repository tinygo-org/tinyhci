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

func (build Build) pendingCheckSuite() {
	log.Printf("Github check suite pending for %s\n", build.sha)
	for _, board := range boards {
		build.pendingCheckRun(board.target)
	}
}

func (build Build) pendingCheckRun(target string) {
	log.Printf("Github check run pending on board %s for %s\n", target, build.sha)
	opts := github.CreateCheckRunOptions{
		Name:    targetName(target),
		HeadSHA: build.sha,
	}
	cr, _, err := client.Checks.CreateCheckRun(context.Background(), ghorg, ghrepo, opts)
	if err != nil {
		log.Println(err)
	}
	build.runs[target] = cr
}

func (build Build) startCheckSuite() {
	log.Printf("Github check suite starting for %s\n", build.sha)
	for _, run := range build.runs {
		build.startCheckRun(run.GetName())
	}
}

func (build Build) startCheckRun(target string) {
	log.Printf("Github check run starting on board %s for %s\n", target, build.sha)
	status := "in_progress"
	if run, ok := build.runs[target]; ok {
		opts := github.UpdateCheckRunOptions{
			Name:   targetName(target),
			Status: &status,
		}
		cr, _, err := client.Checks.UpdateCheckRun(context.Background(), ghorg, ghrepo, *run.ID, opts)
		if err != nil {
			log.Println(err)
		}
		build.runs[target] = cr
	}
}

func (build Build) passCheckRun(target, output string) {
	log.Printf("Github check run passed on board %s for %s\n", target, build.sha)
	title := targetName(target)
	summary := "Hardware CI tests have passed."
	status := "completed"
	conclusion := "success"
	timestamp := github.Timestamp{Time: time.Now()}
	if run, ok := build.runs[target]; ok {
		ro := github.CheckRunOutput{
			Title:   &title,
			Summary: &summary,
			Text:    &output,
		}

		opts := github.UpdateCheckRunOptions{
			Name:        targetName(target),
			Status:      &status,
			Conclusion:  &conclusion,
			CompletedAt: &timestamp,
			Output:      &ro,
		}
		_, _, err := client.Checks.UpdateCheckRun(context.Background(), ghorg, ghrepo, *run.ID, opts)
		if err != nil {
			log.Println(err)
		}
		delete(build.runs, target)
	}
}

func (build Build) failCheckSuite(output string) {
	log.Printf("Github check suite failed for %s\n", build.sha)
	for _, run := range build.runs {
		build.failCheckRun(run.GetName(), output)
	}
}

func (build Build) failCheckRun(target, output string) {
	log.Printf("Github check run failed on board %s for %s\n", target, build.sha)
	title := targetName(target)
	summary := "Hardware CI tests have failed."
	status := "completed"
	conclusion := "failure"
	timestamp := github.Timestamp{Time: time.Now()}
	if run, ok := build.runs[target]; ok {
		ro := github.CheckRunOutput{
			Title:   &title,
			Summary: &summary,
			Text:    &output,
		}

		opts := github.UpdateCheckRunOptions{
			Name:        targetName(target),
			Status:      &status,
			Conclusion:  &conclusion,
			CompletedAt: &timestamp,
			Output:      &ro,
		}
		_, _, err := client.Checks.UpdateCheckRun(context.Background(), ghorg, ghrepo, *run.ID, opts)
		if err != nil {
			log.Println(err)
		}
		delete(build.runs, target)
	}
}

func targetName(target string) string {
	return target
}
