package main

import (
	"log"
	"time"

	"github.com/google/go-github/v31/github"
)

// Build is a specific build to be tested.
type Build struct {
	pendingCI bool
	started   time.Time
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
	if !board.enabled {
		log.Printf("Board %s has been disabled, so passing.\n", board.displayname)
		build.passCheckRun(board.target, "Board disabled in TinyHCI.")
	}

	log.Printf("Flashing board %s\n", board.displayname)
	fout, err := board.flash(build.sha)
	if err != nil {
		log.Println(err)
		log.Println(fout)
		build.failCheckRun(board.target, flashout(fout))
		return
	}

	time.Sleep(board.resetpause)

	log.Printf("Running tests on board %s\n", board.displayname)
	out, err := board.test()
	if err != nil {
		log.Println(err)
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
