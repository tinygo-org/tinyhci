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
	fout, err := board.flash(build.sha)
	if err != nil {
		log.Println(err)
		log.Println(fout)
		build.failCheckRun(board.target, flashout(fout))
		return
	}

	time.Sleep(5 * time.Second)

	log.Printf("Running tests on board %s\n", board.displayname)
	out, err := board.test()
	if err != nil {
		log.Println(err)
		log.Println(out)
		build.failCheckRun(board.target, flashout(fout)+testsout(out))
		return
	}

	build.passCheckRun(board.target, flashout(fout)+testsout(out))
}

func flashout(out string) string {
	return "## Flash\n\n```\n" +
		out +
		"\n```\n\n"
}

func testsout(out string) string {
	return "## Tests\n\n" +
		out +
		"\n\n"
}
