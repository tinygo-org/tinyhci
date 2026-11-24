package main

import (
	"log"
	"sync"
	"time"

	"github.com/google/go-github/v84/github"
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
		return
	}

	log.Printf("Flashing board %s\n", board.displayname)
	fout, err := board.flash(build.sha)
	if err != nil {
		log.Println(err)
		log.Println(fout)
		build.failCheckRun(board.target, boardHeading(board)+flashout(fout))
		return
	}

	time.Sleep(board.resetpause)

	log.Printf("Running tests on board %s\n", board.displayname)
	out, err := board.test()
	if err != nil {
		log.Println(err)
		build.failCheckRun(board.target, boardHeading(board)+flashout(fout)+testsout(out))
		return
	}

	build.passCheckRun(board.target, boardHeading(board)+flashout(fout)+testsout(out))
}

// compileBoards compiles the tests for all boards at the same time.
// Errors are ignored here because the flash step reports them.
func (build Build) compileBoards() {
	var wg sync.WaitGroup
	for target := range build.runs {
		board := GetBoard(target)
		if board == nil || !board.enabled {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("Compiling tests for board %s\n", board.displayname)
			if out, err := board.compile(build.sha); err != nil {
				log.Printf("Compile failed on board %s: %v\n%s\n", board.displayname, err, out)
			}
		}()
	}
	wg.Wait()
}

func boardHeading(board *Board) string {
	heading := "## " + board.displayname + "\n\n"
	if board.image != "" {
		heading = "![board image](https://raw.githubusercontent.com/tinygo-org/tinyhci/refs/heads/main/images/boards/" + board.image + ")\n\n" + heading
	}

	return heading
}

func flashout(out string) string {
	return "## Flash\n\n```\n" +
		out +
		"\n```\n\n"
}

func testsout(out string) string {
	return "## Tests\n\n```\n" +
		out +
		"\n```\n\n"
}
