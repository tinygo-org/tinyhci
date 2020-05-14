package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/metanerd/go-circleci"
)

// CIBuildInfo is info from CircleCI webhook.
type CIBuildInfo struct {
	BuildNum    string `json:"build_num"`
	Branch      string `json:"branch"`
	Username    string `json:"username"`
	Job         string `json:"job"`
	BuildUrl    string `json:"build_url"`
	VcsRevision string `json:"vcs_revision"`
	RepoName    string `json:"reponame"`
	WorkflowId  string `json:"workflow_id"`
	WorkflowUrl string `json:"workflow_url"`
	PullRequest string `json:"pull_request"`
	User        string `json:"user"`
	ApiLink     string `json:"api_link"`
	Status      string `json:"status"`
}

func parseBuildInfo(r *http.Request) (*CIBuildInfo, error) {
	// Read body
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, err
	}

	// Unmarshal
	var bi CIBuildInfo
	err = json.Unmarshal(b, &bi)
	if err != nil {
		return nil, err
	}

	return &bi, nil
}

func getTinygoBinaryURL(buildNum string) (string, error) {
	client := &circleci.Client{}
	bn, err := strconv.Atoi(buildNum)
	if err != nil {
		return "", errors.New("invalid buildnum: " + buildNum)
	}
	artifacts, _ := client.ListBuildArtifacts("github", "tinygo-org", "tinygo", bn)

	for _, a := range artifacts {
		// we're looking for the .deb file
		if a.Path == "tmp/tinygo.linux-amd64.tar.gz" {
			return a.URL, nil
		}
	}
	return "", errors.New("cannot find DEB file")
}
