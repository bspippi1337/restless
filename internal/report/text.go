package report

import (
	"fmt"
	"io"

	"restless/internal/core"
)

type TextOptions struct {
	ShowLatency bool
}

func WriteText(w io.Writer, result core.VerificationResult, opts TextOptions) error {
	for _, r := range result.Results {
		status := formatStatus(r.Status)

		if opts.ShowLatency {
			fmt.Fprintf(w, "%-5s %-6s %-30s %dms",
				status,
				r.Endpoint.Method,
				r.Endpoint.Path,
				r.Latency.Milliseconds(),
			)
		} else {
			fmt.Fprintf(w, "%-5s %-6s %-30s",
				status,
				r.Endpoint.Method,
				r.Endpoint.Path,
			)
		}

		if len(r.Issues) > 0 {
			fmt.Fprintf(w, " %s", r.Issues[0].Message)
		}

		fmt.Fprintln(w)
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Summary:")
	fmt.Fprintf(w, "  OK:   %d\n", result.Summary.OK)
	fmt.Fprintf(w, "  WARN: %d\n", result.Summary.Warn)
	fmt.Fprintf(w, "  FAIL: %d\n", result.Summary.Fail)

	return nil
}

func formatStatus(s core.Status) string {
	switch s {
	case core.StatusOK:
		return "OK"
	case core.StatusWarn:
		return "WARN"
	case core.StatusFail:
		return "FAIL"
	default:
		return "FAIL"
	}
}
