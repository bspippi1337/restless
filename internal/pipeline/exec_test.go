package pipeline

import (
	"strings"
	"testing"

	"github.com/bspippi1337/restless/internal/events"
)

func TestRunCapturesStdoutAndExitCode(t *testing.T) {
	ev := events.New("test", "filesystem", "example.txt")
	res := Run(ev, "printf restless")

	if res.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d stderr=%q", res.ExitCode, res.Stderr)
	}
	if strings.TrimSpace(res.Stdout) != "restless" {
		t.Fatalf("expected stdout restless, got %q", res.Stdout)
	}
	if res.Command != "printf restless" {
		t.Fatalf("command mismatch: %q", res.Command)
	}
	if res.DurationMS < 0 {
		t.Fatalf("duration must not be negative")
	}
}

func TestRunReportsFailure(t *testing.T) {
	ev := events.New("test", "filesystem", "example.txt")
	res := Run(ev, "exit 7")

	if res.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code")
	}
}
