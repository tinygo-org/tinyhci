package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var (
	URL          = "https://circleci.com/api/v2"
	PROJECT_SLUG = "/gh/" + ghorg + "/" + ghrepo + "/"
)

type WorkflowRuns struct {
	Items         []WorkflowRun `json:"items"`
	NextPageToken string        `json:"next_page_token"`
}

type WorkflowRun struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	Duration    int       `json:"duration"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	StoppedAt   time.Time `json:"stopped_at"`
	CreditsUsed int       `json:"credits_used"`
	Branch      string    `json:"branch"`
}

type WorkflowJobs struct {
	Items         []WorkflowJob `json:"items"`
	NextPageToken string        `json:"next_page_token"`
}

type WorkflowJob struct {
	CanceledBy  string    `json:"canceled_by"`
	JobNumber   int       `json:"job_number"`
	ID          string    `json:"id"`
	StartedAt   time.Time `json:"started_at"`
	Name        string    `json:"name"`
	ApprovedBy  string    `json:"approved_by"`
	ProjectSlug string    `json:"project_slug"`
	Status      string    `json:"status"`
	Type        string    `json:"type"`
	StoppedAt   time.Time `json:"stopped_at"`
}

type JobArtifacts struct {
	Items         []JobArtifact `json:"items"`
	NextPageToken string        `json:"next_page_token"`
}

type JobArtifact struct {
	Path      string `json:"path"`
	NodeIndex int    `json:"node_index"`
	Url       string `json:"url"`
}

type Workflow struct {
	PipelineID     string    `json:"pipeline_id"`
	CanceledBy     string    `json:"canceled_by"`
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	ProjectSlug    string    `json:"project_slug"`
	ErroredBy      string    `json:"errored_by"`
	Tag            string    `json:"tag"`
	Status         string    `json:"status"`
	StartedBy      string    `json:"started_by"`
	CreatedAt      time.Time `json:"created_at"`
	PipelineNumber int       `json:"pipeline_number"`
	StoppedAt      time.Time `json:"stopped_at"`
}

type Pipeline struct {
	ID          string    `json:"id"`
	ProjectSlug string    `json:"project_slug"`
	UpdatedAt   time.Time `json:"updated_at"`
	Number      int       `json:"number"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	Vcs         VCS       `json:"vcs"`
}

type VCS struct {
	ProviderName        string `json:"provider_name"`
	TargetRepositoryUrl string `json:"target_repository_url"`
	Branch              string `json:"branch"`
	ReviewID            string `json:"review_id"`
	ReviewUrl           string `json:"review_url"`
	Revision            string `json:"revision"`
	Tag                 string `json:"tag"`
	OriginRepositoryUrl string `json:"origin_repository_url"`
}

func getTinygoBinaryURL(buildNum int) (string, error) {
	if useCurrentBinaryRelease {
		return "using current TinyGo binary release", nil
	}

	artifacts, err := getJobArtifacts(buildNum)
	if err != nil {
		return "", err
	}

	for _, a := range artifacts.Items {
		// we're looking for the .tar.gz file
		if a.Path == "tmp/tinygo.linux-amd64.tar.gz" {
			return a.Url, nil
		}
	}
	return "", fmt.Errorf("cannot find build artifact file for build %d", buildNum)
}

func getCIBuildNumFromSHA(sha string) (int, error) {
	if useCurrentBinaryRelease {
		return -1, nil
	}

	wfr, err := getWorkflowRuns(time.Now().Add(-time.Hour * 12))
	if err != nil {
		return -1, err
	}

	for _, v := range wfr.Items {
		if v.Status != "success" {
			continue
		}

		wf, err := getWorkflow(v.ID)
		if err != nil {
			return -1, err
		}

		pl, err := getPipeline(wf.PipelineID)
		if err != nil {
			return -1, err
		}

		if pl.Vcs.Revision != sha {
			continue
		}

		jobs, err := getWorkflowJobs(wf.ID)
		if err != nil {
			return -1, err
		}

		for _, job := range jobs.Items {
			if !(job.Name == "build-linux" && job.Status == "success") {
				continue
			}

			return job.JobNumber, nil
		}
	}

	return -1, fmt.Errorf("cannot find TinyGo build for %s", sha)
}

func getMostRecentCIBuildNumAfterStart(sha string, start time.Time) (int, error) {
	if useCurrentBinaryRelease {
		return -1, nil
	}

	wfr, err := getWorkflowRuns(start)
	if err != nil {
		return -1, err
	}

	for _, v := range wfr.Items {
		if v.Status == "success" {
			wf, err := getWorkflow(v.ID)
			if err != nil {
				return -1, err
			}

			pl, err := getPipeline(wf.PipelineID)
			if err != nil {
				return -1, err
			}

			if pl.Vcs.Revision == sha {
				jobs, err := getWorkflowJobs(wf.ID)
				if err != nil {
					return -1, err
				}

				for _, job := range jobs.Items {
					if !(job.Name == "build-linux" && job.Status == "success") {
						continue
					}

					return job.JobNumber, nil
				}
			}
		}
	}

	return -1, fmt.Errorf("cannot find recent TinyGo build for %s", sha)
}

func getRecentSuccessfulCIBuilds() ([]Pipeline, error) {
	pls := make([]Pipeline, 0)

	wfr, err := getWorkflowRuns(time.Now().Add(-time.Hour * 12))
	if err != nil {
		return pls, err
	}

	for _, v := range wfr.Items {
		if v.Status == "success" {
			jobs, err := getWorkflowJobs(v.ID)
			if err != nil {
				return pls, err
			}

			for _, j := range jobs.Items {
				if !(j.Name == "build-linux" && j.Status == "success") {
					continue
				}

				wf, err := getWorkflow(v.ID)
				if err != nil {
					return pls, err
				}

				pl, err := getPipeline(wf.PipelineID)
				if err != nil {
					return pls, err
				}
				pls = append(pls, pl)
			}
		}
	}

	return pls, nil
}

func getWorkflowRuns(start time.Time) (WorkflowRuns, error) {
	var runs WorkflowRuns
	url := URL + "/insights" + PROJECT_SLUG + "workflows/test-all?all-branches=true&start-date=" + start.Format(time.RFC3339)
	fmt.Println("getWorkflowRuns: ", url)
	body, err := callCircleCIAPI(url)
	if err != nil {
		return runs, err
	}

	json.Unmarshal([]byte(body), &runs)
	return runs, nil
}

func getWorkflowRunsBranch(branch string, start time.Time) (WorkflowRuns, error) {
	var runs WorkflowRuns
	url := URL + "/insights" + PROJECT_SLUG + "workflows/test-all?all-branches=false&" +
		"&branch=" + branch +
		"&start-date=" + start.Format(time.RFC3339)

	body, err := callCircleCIAPI(url)
	if err != nil {
		return runs, err
	}

	json.Unmarshal([]byte(body), &runs)
	return runs, nil
}

func getWorkflow(id string) (Workflow, error) {
	var workflow Workflow
	url := URL + "/workflow/" + id
	body, err := callCircleCIAPI(url)
	if err != nil {
		return workflow, err
	}

	json.Unmarshal([]byte(body), &workflow)
	return workflow, nil
}

func getWorkflowJobs(id string) (WorkflowJobs, error) {
	var jobs WorkflowJobs
	url := URL + "/workflow/" + id + "/job"
	body, err := callCircleCIAPI(url)
	if err != nil {
		return jobs, err
	}

	json.Unmarshal([]byte(body), &jobs)
	return jobs, nil
}

func getJobArtifacts(id int) (JobArtifacts, error) {
	var artifacts JobArtifacts
	url := URL + "/project" + PROJECT_SLUG + strconv.Itoa(id) + "/artifacts"
	body, err := callCircleCIAPI(url)
	if err != nil {
		return artifacts, err
	}

	json.Unmarshal([]byte(body), &artifacts)
	return artifacts, nil
}

func getPipeline(id string) (Pipeline, error) {
	var pipeline Pipeline
	url := URL + "/pipeline/" + id
	body, err := callCircleCIAPI(url)
	if err != nil {
		return pipeline, err
	}

	json.Unmarshal([]byte(body), &pipeline)
	return pipeline, nil
}

func callCircleCIAPI(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Circle-Token", citoken)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

