package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/metanerd/go-circleci"
)

// BuildInfo is info from CircleCI webhook.
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

func getTinygoBinaryURL(buildNum string) (string, error) {
	client := &circleci.Client{} //Token: "YOUR TOKEN"} // Token not required to query info for public projects
	bn, err := strconv.Atoi(buildNum)
	if err != nil {
		return "", errors.New("invalid buildnum: " + buildNum)
	}
	artifacts, _ := client.ListBuildArtifacts("github", "tinygo-org", "tinygo", bn)

	for _, a := range artifacts {
		// we're looking for the .deb file
		if a.Path == "tmp/tinygo_amd64.deb" {
			return a.URL, nil
		}
	}
	return "", errors.New("cannot find DEB file")
}

func downloadFile(filepath string, url string) (err error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
