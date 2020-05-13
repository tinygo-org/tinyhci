package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// BuildInfo is info from CircleCI.
type BuildInfo struct {
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

func parseBuildInfo(r *http.Request) (*BuildInfo, error) {
	// Read body
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, err
	}

	// Unmarshal
	var bi BuildInfo
	err = json.Unmarshal(b, &bi)
	if err != nil {
		return nil, err
	}

	return &bi, nil
}
