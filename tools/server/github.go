package main

import (
	"log"
	"net/http"

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
}

func startCheckRun(sha string) {
	log.Printf("Github check run starting for %s\n", sha)
}

func passCheckRun(sha string) {
	log.Printf("Github check run pass for commit %s\n", sha)
}

func failCheckRun(sha string) {
	log.Printf("Github check run fail for commit %s\n", sha)
}
