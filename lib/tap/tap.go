// package tap provides a simple TAP (Test Anything Protocol) implementation for Go.
// It is designed to be used in test code, and provides methods to generate
// TAP output indicating test results.
package tap

import (
	"strconv"
	"strings"
)

// Tester is a type to encapsulate test state.  Methods on this type generate TAP
// output.
type Tester struct {
	nextTestNumber int

	// TODO toggles the TODO directive for Ok, Fail, Pass, and similar.
	TODO bool

	print func(messages ...any)
}

func New() *Tester {
	return &Tester{
		nextTestNumber: 1,
		print:          printer,
	}
}

// Header displays a TAP header including version number and expected
// number of tests to run.
func (t *Tester) Header(testCount int) {
	t.print("TAP version 13")
	if testCount > 0 {
		t.print("1.." + strconv.Itoa(testCount))
	}
}

// Ok generates TAP output indicating whether a test passed or failed.
func (t *Tester) Ok(test bool, description string) {
	// did the test pass or not?
	ok := "ok"
	if !test {
		ok = "not ok"
	}

	if t.TODO {
		t.print(ok, t.nextTestNumber, "# TODO", description)
	} else {
		t.print(ok, t.nextTestNumber, "-", description)
	}
	t.nextTestNumber++
}

// Fail indicates that a test has failed.  This is typically only used when the
// logic is too complex to fit naturally into an Ok() call.
func (t *Tester) Fail(description string) {
	t.Ok(false, description)
}

// Pass indicates that a test has passed.  This is typically only used when the
// logic is too complex to fit naturally into an Ok() call.
func (t *Tester) Pass(description string) {
	t.Ok(true, description)
}

// Skip indicates that a test has been skipped.
func (t *Tester) Skip(count int, description string) {
	for i := 0; i < count; i++ {
		t.print("ok", t.nextTestNumber, "# SKIP", description)
		t.nextTestNumber++
	}
}

// Diagnostic generates a diagnostic from the message,
// which may span multiple lines.
func (t *Tester) Diagnostic(message string) {
	t.print("#", escapeNewlines(message))
}

func printer(messages ...any) {
	for i, m := range messages {
		if i > 0 {
			print(" ")
		}
		print(m)
	}
	print("\n")
}

func escapeNewlines(s string) string {
	return strings.Replace(strings.TrimRight(s, "\n"), "\n", "\n# ", -1)
}
