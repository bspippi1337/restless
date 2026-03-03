package main

import (
	"fmt"
	"os"
	"time"

	"github.com/bspippi1337/restless/internal/core"
	"github.com/bspippi1337/restless/internal/httpx"
	"github.com/bspippi1337/restless/internal/report"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: restless <command>")
		os.Exit(3)
	}

	switch os.Args[1] {
	case "verify":
		runVerify()
	default:
		fmt.Println("unknown command")
		os.Exit(3)
	}
}

func runVerify() {
	base := "https://api.github.com"

	exec := httpx.NewExecutor(10 * time.Second)
	agg := core.NewAggregator()

	endpoints := []core.Endpoint{
		{Method: "GET", Path: "/users/octocat"},
		{Method: "GET", Path: "/repos/octocat/Hello-World"},
	}

	for _, ep := range endpoints {
		url := base + ep.Path

		resp, err := exec.Do(ep.Method, url)
		if err != nil {
			agg.Add(core.EndpointResult{
				Endpoint: ep,
				Status:   core.StatusFail,
				Issues: []core.VerificationIssue{
					{Message: err.Error()},
				},
			})
			continue
		}

		status := core.StatusOK
		if resp.StatusCode >= 500 {
			status = core.StatusFail
		} else if resp.StatusCode >= 400 {
			status = core.StatusWarn
		}

		agg.Add(core.EndpointResult{
			Endpoint:   ep,
			Status:     status,
			HTTPStatus: resp.StatusCode,
			Latency:    resp.Latency,
		})
	}

	result := agg.Build("dev-spec-hash", base)

	opts := report.TextOptions{ShowLatency: true}

	if err := report.WriteText(os.Stdout, result, opts); err != nil {
		fmt.Println("report error:", err)
		os.Exit(3)
	}

	if result.Summary.Fail > 0 {
		os.Exit(2)
	}
	if result.Summary.Warn > 0 {
		os.Exit(1)
	}
	os.Exit(0)
}
