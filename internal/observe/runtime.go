package observe

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/bspippi1337/restless/internal/events"
)

type Execution struct {
	Event      events.Event `json:"event"`
	Command    string       `json:"command"`
	DurationMS int64        `json:"duration_ms"`
	ExitCode   int          `json:"exit_code"`
	Stdout     string       `json:"stdout,omitempty"`
	Stderr     string       `json:"stderr,omitempty"`
	StartedAt  time.Time    `json:"started_at"`
	FinishedAt time.Time    `json:"finished_at"`
}

func PrintHuman(e Execution) {
	fmt.Printf("[%s] %s -> %s (%dms, exit=%d)\n",
		e.Event.Kind,
		e.Event.Path,
		e.Command,
		e.DurationMS,
		e.ExitCode,
	)
}

func PrintJSON(e Execution) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(e)
}
