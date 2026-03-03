package main

import (
	"fmt"
	"os"
)

var version = "dev"
import (

var version = "dev"





	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bspippi1337/restless/internal/core"
	"github.com/bspippi1337/restless/internal/history"
	"github.com/bspippi1337/restless/internal/httpx"
	"github.com/bspippi1337/restless/internal/insight"
	"github.com/bspippi1337/restless/internal/openapi"
	"github.com/bspippi1337/restless/internal/probe"
	"github.com/bspippi1337/restless/internal/report"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: restless <command>")
		os.Exit(3)
	}

	switch os.Args[1] {

case "version":
	fmt.Println(version)
	return


	case "verify":
		runVerify(os.Args[2:])

	case "map":

	case "scan":

	case "inspect":
		runInspect(os.Args[2:])

		runScan(os.Args[2:])

		runMap(os.Args[2:])

	default:
		fmt.Println("unknown command")
		os.Exit(3)
	}
}

func runVerify(args []string) {

	jsonMode := false
	showLatency := false
	enableInsights := false
	base := "https://api.github.com"
	workers := 100
	specFile := ""

	for i := 0; i < len(args); i++ {

		a := args[i]

		switch a {

		case "--json":
			jsonMode = true

		case "--latency":
			showLatency = true

		case "--insights":
			enableInsights = true

		case "--base":
			if i+1 < len(args) {
				base = strings.TrimRight(args[i+1], "/")
				i++
			}

		case "--workers":
			if i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil && n > 0 {
					workers = n
				}
				i++
			}

		case "--spec":
			if i+1 < len(args) {
				specFile = args[i+1]
				i++
			}
		}
	}

	exec := httpx.NewExecutor(10 * time.Second)
	agg := core.NewAggregator()

	var endpoints []core.Endpoint

	if specFile != "" {

		eps, err := openapi.Load(specFile)
		if err != nil {
			fmt.Println("spec load error:", err)
			os.Exit(3)
		}

		endpoints = probe.Plan(eps)

	} else {

		endpoints = []core.Endpoint{
			{Method: "GET", Path: "/users/octocat"},
			{Method: "GET", Path: "/repos/octocat/Hello-World"},
		}
	}

	var meta core.Meta

	jobs := make(chan core.Endpoint)
	results := make(chan core.EndpointResult)

	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {

		wg.Add(1)

		go func() {

			defer wg.Done()

			for ep := range jobs {

				req := probe.Build(base, ep)

				var body []byte

				switch ep.Method {
				case "POST", "PUT", "PATCH":
					body = probe.SimpleJSONBody()
				}

				resp, err := exec.Do(ep.Method, req.URL, body)

				if err != nil {

					results <- core.EndpointResult{
						Endpoint: ep,
						Status:   core.StatusFail,
						Issues: []core.VerificationIssue{
							{Message: err.Error()},
						},
					}

					continue
				}

				meta.RateLimitRemaining = resp.RateLimitRemaining
				meta.RateLimitReset = resp.RateLimitReset

				status := core.StatusOK

				if resp.StatusCode >= 500 {
					status = core.StatusFail
				} else if resp.StatusCode >= 400 {
					status = core.StatusWarn
				}

				results <- core.EndpointResult{
					Endpoint:   ep,
					Status:     status,
					HTTPStatus: resp.StatusCode,
					Latency:    resp.Latency,
				}
			}
		}()
	}

	go func() {

		for _, ep := range endpoints {
			jobs <- ep
		}

		close(jobs)

	}()

	go func() {

		wg.Wait()
		close(results)

	}()

	for r := range results {
		agg.Add(r)
	}

	agg.SetMeta(meta)

	result := agg.Build("dev-spec-hash", base)

	if enableInsights {

		result.Insights = insight.Analyze(result.Results)

		if prev, err := history.Load(); err == nil {

			result.Insights = append(
				result.Insights,
				insight.DriftInsights(prev.Results, result.Results)...,
			)
		}
	}

	_ = history.Save(result)

	if jsonMode {

		report.WriteJSON(os.Stdout, result, report.JSONOptions{Pretty: true})

	} else {

		report.WriteText(os.Stdout, result, report.TextOptions{ShowLatency: showLatency})
	}

	if enableInsights && !jsonMode {

		for _, i := range result.Insights {
			fmt.Println("Insight:", i.Message)
		}
	}

	if result.Summary.Fail > 0 {
		os.Exit(2)
	}

	if result.Summary.Warn > 0 {
		os.Exit(1)
	}

	os.Exit(0)
}
