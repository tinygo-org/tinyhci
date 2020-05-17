package main

import (
	"log"
	"time"

	"github.com/google/go-github/v31/github"
)

// Build is a specific build to be tested.
type Build struct {
	binaryURL string
	sha       string
	suite     *github.CheckSuite

	// runs are all of the checkruns for this build.
	// key is the target.
	runs map[string]*github.CheckRun
}

// NewBuild returns a new Build.
func NewBuild(sha string) *Build {
	return &Build{
		sha:  sha,
		runs: make(map[string]*github.CheckRun),
	}
}

func (build Build) processBoardRun(board *Board) {
	log.Printf("Flashing board %s\n", board.displayname)
	flashout, err := board.flash(build.sha)
	if err != nil {
		log.Println(err)
		log.Println(flashout)
		build.failCheckRun(board.target, flashout)
		return
	}

	time.Sleep(2 * time.Second)

	log.Printf("Running tests on board %s\n", board.displayname)
	out, err := board.test()
	if err != nil {
		log.Println(err)
		log.Println(out)
		build.failCheckRun(board.target, out)
		return
	}
	build.passCheckRun(board.target, out)
}
