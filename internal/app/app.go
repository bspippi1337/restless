package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/bspippi1337/restless/internal/core/discovery"
	"github.com/bspippi1337/restless/internal/core/doctor"
	"github.com/bspippi1337/restless/internal/tui"
)

const version = "v0.2.0-alpha"

func Run(args []string, stdin, stdout, stderr *os.File) int {
	if len(args) > 0 {
		switch args[0] {
		case "help":
			fmt.Fprintln(stdout, helpText())
			return 0
		case "discover":
			return cmdDiscover(args[1:], stdout, stderr)
		case "doctor":
			return cmdDoctor(args[1:], stdout, stderr)
		}
	}

	fs := flag.NewFlagSet("restless", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var mode string
	fs.StringVar(&mode, "mode", "auto", "Mode: auto|tui")
	fs.StringVar(&mode, "m", "auto", "Alias for --mode")
	var quiet bool
	fs.BoolVar(&quiet, "quiet", false, "Disable animations and extra UI flair")
	var showVersion bool
	fs.BoolVar(&showVersion, "version", false, "Print version and exit")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if showVersion {
		fmt.Fprintln(stdout, "restless", version)
		return 0
	}

	if mode == "auto" || mode == "tui" {
		if err := tui.Run(stdin, stdout, quiet); err != nil {
			fmt.Fprintln(stderr, "restless:", err)
			return 1
		}
		return 0
	}

	fmt.Fprintln(stderr, "Unknown mode. v2 supports: --mode tui")
	return 2
}

func cmdDiscover(args []string, stdout, stderr *os.File) int {
	fs := flag.NewFlagSet("restless discover", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var asJSON bool
	fs.BoolVar(&asJSON, "json", false, "Print findings as JSON")
	var verify bool
	fs.BoolVar(&verify, "verify", true, "Safe verify endpoints (GET/HEAD/OPTIONS)")
	var seconds int
	fs.IntVar(&seconds, "seconds", 15, "Time budget in seconds")
	var pages int
	fs.IntVar(&pages, "pages", 6, "Doc pages budget")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	domain := ""
	if fs.NArg() > 0 {
		domain = strings.TrimSpace(fs.Arg(0))
	}
	if domain == "" {
		fmt.Fprintln(stderr, "Usage: restless discover <domain> [--json]")
		return 2
	}

	find, err := discovery.DiscoverDomain(domain, discovery.Options{
		BudgetSeconds: seconds,
		BudgetPages:   pages,
		Verify:        verify,
		Fuzz:          true,
	})
	if err != nil {
		fmt.Fprintln(stderr, "discover:", err)
		return 1
	}

	if asJSON {
		b, _ := json.MarshalIndent(find, "", "  ")
		fmt.Fprintln(stdout, string(b))
		return 0
	}

	fmt.Fprintf(stdout, "Base: %s\n", find.BaseURL)
	if len(find.DocURLs) > 0 {
		fmt.Fprintln(stdout, "Docs:")
		for _, u := range find.DocURLs {
			fmt.Fprintf(stdout, "  - %s\n", u)
		}
	}
	fmt.Fprintf(stdout, "Endpoints (%d):\n", len(find.Endpoints))
	max := 20
	if len(find.Endpoints) < max {
		max = len(find.Endpoints)
	}
	for i := 0; i < max; i++ {
		e := find.Endpoints[i]
		fmt.Fprintf(stdout, "  %s %s\n", e.Method, e.Path)
	}
	if len(find.Notes) > 0 {
		fmt.Fprintln(stdout, "Notes:")
		for _, n := range find.Notes {
			fmt.Fprintf(stdout, "  - %s\n", n)
		}
	}
	return 0
}

func cmdDoctor(args []string, stdout, stderr *os.File) int {
	fs := flag.NewFlagSet("restless doctor", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var rootDir string
	fs.StringVar(&rootDir, "root", ".", "Project root to clean (default: current dir)")
	var dry bool
	fs.BoolVar(&dry, "dry-run", false, "Print actions without deleting")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	rep, err := doctor.Run(rootDir, dry)
	if err != nil {
		fmt.Fprintln(stderr, "doctor:", err)
		return 1
	}
	fmt.Fprintln(stdout, rep)
	return 0
}

func helpText() string {
	return `restless — CLI-first API client (domain-first discovery)

Usage:
  restless [flags]
  restless help
  restless discover <domain> [--json]
  restless doctor

Flags:
  --mode tui        Run terminal UI (default auto)
  --quiet           Disable animations and extra UI flair
  --version         Print version and exit
  --help            Show this help

Quick start:
  restless
  # In "Connect & Discover", type a domain and press Ctrl+D

Tips:
  - Ctrl+D runs discovery in the wizard
  - Press ? to open the in-app manual
  - 'restless help' prints a guide pointer

Docs:
  docs/RFC-0002-domain-first-discovery.md
`
}
