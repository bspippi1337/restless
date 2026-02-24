package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bspippi1337/restless/internal/core/app"
	"github.com/bspippi1337/restless/internal/core/types"
	"github.com/bspippi1337/restless/internal/modules/bench"
	"github.com/bspippi1337/restless/internal/modules/export"
	"github.com/bspippi1337/restless/internal/modules/openapi"
	"github.com/bspippi1337/restless/internal/modules/session"
)

func main() {
	var (
		method = flag.String("X", "GET", "HTTP method")
		url    = flag.String("url", "", "Request URL")
		body   = flag.String("d", "", "Body string")
		hdrK   = flag.String("Hk", "", "Header key (single)")
		hdrV   = flag.String("Hv", "", "Header value (single)")
		setVar = flag.String("set", "", "Set session var: key=value")
		doBench = flag.Bool("bench", false, "Run bench mode")
		c       = flag.Int("c", 10, "Bench concurrency")
		dur     = flag.Duration("dur", 5*time.Second, "Bench duration")
		save    = flag.String("save", "", "Save json artifact name")
	)
	flag.Parse()

	// Modules
	sess := session.New()
	_ = openapi.New() // not used yet here, but wired
	_ = export.New()
	_ = bench.New()

	mods := []app.Module{
		sess,
		openapi.New(),
		export.New(),
		bench.New(),
	}

	a, err := app.New(mods)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if *setVar != "" {
		k, v, ok := splitKV(*setVar)
		if !ok {
			fmt.Fprintln(os.Stderr, "invalid -set, want key=value")
			os.Exit(2)
		}
		sess.Set(k, v)
	}

	if *url == "" {
		fmt.Fprintln(os.Stderr, "missing -url")
		flag.Usage()
		os.Exit(2)
	}

	h := http.Header{}
	if *hdrK != "" {
		h.Add(*hdrK, *hdrV)
	}

	req := types.Request{
		Method:  *method,
		URL:     *url,
		Headers: h,
		Body:    []byte(*body),
	}

	ctx := context.Background()

	if *doBench {
		r, err := bench.Run(ctx, aRunner(a), bench.Config{
			Concurrency: *c,
			Duration:    *dur,
			Request:     req,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "bench error:", err)
			os.Exit(1)
		}
		fmt.Printf("bench: total=%d errors=%d dur_ms=%d p50=%d p95=%d p99=%d\n",
			r.TotalRequests, r.Errors, r.DurationMs, r.P50Ms, r.P95Ms, r.P99Ms)
		return
	}

	resp, err := a.RunOnce(ctx, req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "request error:", err)
		os.Exit(1)
	}

	fmt.Printf("status: %d (dur=%dms)\n", resp.StatusCode, resp.DurationMs)
	fmt.Printf("%s\n", string(resp.Body))

	if *save != "" {
		p, err := export.SaveJSONArtifact(*save, resp)
		if err != nil {
			fmt.Fprintln(os.Stderr, "save error:", err)
			os.Exit(1)
		}
		fmt.Printf("saved: %s\n", p)
	}
}

func splitKV(s string) (k, v string, ok bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return s[:i], s[i+1:], true
		}
	}
	return "", "", false
}

// aRunner adapts App.RunOnce into the bench runner interface without exposing internals.
type runOnceAdapter struct{ a *app.App }

func aRunner(a *app.App) *runOnceAdapter { return &runOnceAdapter{a: a} }

func (r *runOnceAdapter) Run(ctx context.Context, req types.Request) (types.Response, error) {
	return r.a.RunOnce(ctx, req)
}
