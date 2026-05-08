package pipeline

import (
	"bytes"
	"os/exec"
	"time"

	"github.com/bspippi1337/restless/internal/events"
	"github.com/bspippi1337/restless/internal/observe"
)

func Run(event events.Event, command string) observe.Execution {
	started := time.Now().UTC()

	cmd := exec.Command("sh", "-c", command)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		exitCode = 1
	}

	finished := time.Now().UTC()

	return observe.Execution{
		Event:      event,
		Command:    command,
		DurationMS: finished.Sub(started).Milliseconds(),
		ExitCode:   exitCode,
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		StartedAt:  started,
		FinishedAt: finished,
	}
}
