package legacy

import "github.com/bspippi1337/restless/internal/teacher"

import (
	"context"
	"flag"
	"fmt"
	"github.com/bspippi1337/restless/internal/version"
	"net/http"
	"os"
	"time"

	"github.com/bspippi1337/restless/internal/core/app"
	"github.com/bspippi1337/restless/internal/core/types"
	"github.com/bspippi1337/restless/internal/diff"
	"github.com/bspippi1337/restless/internal/modules/bench"
	"github.com/bspippi1337/restless/internal/modules/export"
	"github.com/bspippi1337/restless/internal/modules/openapi"
	"github.com/bspippi1337/restless/internal/modules/session"
	"github.com/bspippi1337/restless/internal/snapshot"
	"github.com/bspippi1337/restless/internal/validate"
)

func main() {

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "teacher":
			teacher.Run()
			return
		case "--version":
			fmt.Println("restless", version.Version)
			return
		case "openapi":
			handleOpenAPI(os.Args[2:])
			return
		case "profile":
			handleProfile(os.Args[2:])
			return

		case "snapshot":
			handleSnapshot(os.Args[2:])
			return
		case "diff":
			handleDiff(os.Args[2:])
			return
		case "validate":
			handleValidate(os.Args[2:])
			return
		}
	}

	runRequestMode(os.Args[1:])
}

func runRequestMode(args []string) {
	fs := flag.NewFlagSet("request", flag.ExitOnError)

	method := fs.String("X", "GET", "HTTP method")
	url := fs.String("url", "", "Request URL")
	body := fs.String("d", "", "Body string")
	timeout := fs.Int("timeout", 7, "Timeout in seconds")

	fs.Parse(args)

	if *url == "" {
		fmt.Println("missing -url")
		fs.Usage()
		os.Exit(1)
	}

	sess := session.New()
	mods := []app.Module{
		sess,
		openapi.New(),
		export.New(),
		bench.New(),
	}

	a, err := app.New(mods)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	req := types.Request{
		Method:  *method,
		URL:     *url,
		Headers: http.Header{},
		Body:    []byte(*body),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
	defer cancel()

	resp, err := a.RunOnce(ctx, req)
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	fmt.Printf("status: %d (dur=%dms)\n", resp.StatusCode, resp.DurationMs)
	fmt.Println(string(resp.Body))
}

func handleValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)

	spec := fs.String("spec", "", "Path to OpenAPI spec")
	base := fs.String("base", "", "Base URL")
	strict := fs.Bool("strict", false, "Strict mode")
	jsonOut := fs.Bool("json", false, "JSON output")

	fs.Parse(args)

	if *spec == "" || *base == "" {
		fmt.Println("missing --spec or --base")
		fs.Usage()
		os.Exit(1)
	}

	ctx := context.Background()

	rep, err := validate.Run(ctx, validate.Options{
		SpecPath:   *spec,
		BaseURL:    *base,
		StrictLive: *strict,
	})
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	if *jsonOut {
		validate.PrintJSON(rep, os.Stdout)
	} else {
		validate.PrintHuman(rep, os.Stdout)
	}

	if !rep.OK {
		os.Exit(1)
	}
}

func handleSnapshot(args []string) {
	fs := flag.NewFlagSet("snapshot", flag.ExitOnError)

	spec := fs.String("spec", "", "Path to OpenAPI spec")
	base := fs.String("base", "", "Base URL")
	out := fs.String("out", "snapshot.json", "Output file")
	timeout := fs.Int("timeout", 7, "Timeout in seconds")
	strict := fs.Bool("strict", false, "Strict mode (fail on unexpected status codes)")
	jsonOut := fs.Bool("json", false, "Print snapshot JSON to stdout")

	fs.Parse(args)

	if *spec == "" || *base == "" {
		fmt.Println("missing --spec or --base")
		fs.Usage()
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
	defer cancel()

	rep, err := validate.Run(ctx, validate.Options{
		SpecPath:   *spec,
		BaseURL:    *base,
		Timeout:    time.Duration(*timeout) * time.Second,
		StrictLive: *strict,
	})
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(2)
	}

	snap := snapshot.FromValidateReport(*base, *spec, rep)

	// Always write file (so CI artifacts exist)
	if err := snapshot.WriteJSON(*out, snap); err != nil {
		fmt.Println("ERROR: write snapshot:", err)
		os.Exit(2)
	}

	if *jsonOut {
		b, _ := os.ReadFile(*out)
		fmt.Print(string(b))
	} else {
		snapshot.PrintHuman(os.Stdout, snap)
		fmt.Printf("  wrote: %s\n", *out)
	}

	// Snapshot itself does not fail the build. Use `restless diff` or `restless validate` for gating.
}

func handleDiff(args []string) {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)
	aPath := fs.String("a", "", "Snapshot A path")
	bPath := fs.String("b", "", "Snapshot B path")

	// convenience: allow positional args: restless diff a.json b.json
	fs.Parse(args)
	rest := fs.Args()

	if *aPath == "" && len(rest) >= 1 {
		*aPath = rest[0]
	}
	if *bPath == "" && len(rest) >= 2 {
		*bPath = rest[1]
	}

	if *aPath == "" || *bPath == "" {
		fmt.Println("usage: restless diff A.snap.json B.snap.json")
		fs.Usage()
		os.Exit(2)
	}

	a, err := snapshot.ReadJSON(*aPath)
	if err != nil {
		fmt.Println("ERROR: read A:", err)
		os.Exit(2)
	}
	b, err := snapshot.ReadJSON(*bPath)
	if err != nil {
		fmt.Println("ERROR: read B:", err)
		os.Exit(2)
	}

	r := diff.Compare(a, b)
	diff.PrintHuman(os.Stdout, r)

	if !r.Same {
		os.Exit(1)
	}
}
