package tap

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// helper to create a tester with output captured in a buffer
func newTestTAP() (*Tester, *bytes.Buffer) {
	var buf bytes.Buffer
	tap := &Tester{
		nextTestNumber: 1,
		print: func(messages ...any) {
			for i, m := range messages {
				if i > 0 {
					fmt.Fprint(&buf, " ")
				}
				fmt.Fprint(&buf, m)
			}
			fmt.Fprint(&buf, "\n")
		},
	}
	return tap, &buf
}

func TestHeader(t *testing.T) {
	tap, buf := newTestTAP()
	tap.Header(3)
	expected := "TAP version 13\n1..3\n"
	if buf.String() != expected {
		t.Errorf("Header() output = %q, want %q", buf.String(), expected)
	}
}

func TestOk(t *testing.T) {
	tap, buf := newTestTAP()
	tap.Ok(true, "should pass")
	expected := "ok 1 - should pass\n"
	if buf.String() != expected {
		t.Errorf("Ok(true) output = %q, want %q", buf.String(), expected)
	}
}

func TestNotOk(t *testing.T) {
	tap, buf := newTestTAP()
	tap.Ok(false, "should fail")
	expected := "not ok 1 - should fail\n"
	if buf.String() != expected {
		t.Errorf("Ok(false) output = %q, want %q", buf.String(), expected)
	}
}

func TestOkTODO(t *testing.T) {
	tap, buf := newTestTAP()
	tap.TODO = true
	tap.Ok(true, "pending feature")
	expected := "ok 1 # TODO pending feature\n"
	if buf.String() != expected {
		t.Errorf("Ok(true) with TODO output = %q, want %q", buf.String(), expected)
	}
}

func TestFailAndPass(t *testing.T) {
	tap, buf := newTestTAP()
	tap.Fail("fail desc")
	tap.Pass("pass desc")
	expected := "not ok 1 - fail desc\nok 2 - pass desc\n"
	if buf.String() != expected {
		t.Errorf("Fail/Pass output = %q, want %q", buf.String(), expected)
	}
}

func TestSkip(t *testing.T) {
	tap, buf := newTestTAP()
	tap.Skip(2, "not implemented")
	expected := "ok 1 # SKIP not implemented\nok 2 # SKIP not implemented\n"
	if buf.String() != expected {
		t.Errorf("Skip output = %q, want %q", buf.String(), expected)
	}
}

func TestDiagnosticSingleLine(t *testing.T) {
	tap, buf := newTestTAP()
	tap.Diagnostic("this is a comment")
	expected := "# this is a comment\n"
	if buf.String() != expected {
		t.Errorf("Diagnostic output = %q, want %q", buf.String(), expected)
	}
}

func TestDiagnosticMultiLine(t *testing.T) {
	tap, buf := newTestTAP()
	tap.Diagnostic("line 1\nline 2\nline 3")
	expected := "# line 1\n# line 2\n# line 3\n"
	if buf.String() != expected {
		t.Errorf("Diagnostic multiline output = %q, want %q", buf.String(), expected)
	}
}

func TestNextTestNumberIncrements(t *testing.T) {
	tap, buf := newTestTAP()
	tap.Ok(true, "first")
	tap.Ok(true, "second")
	tap.Ok(false, "third")
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "ok 1") || !strings.HasPrefix(lines[1], "ok 2") || !strings.HasPrefix(lines[2], "not ok 3") {
		t.Errorf("Test numbers did not increment as expected: %v", lines)
	}
}
