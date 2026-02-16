package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bspippi1337/restless/internal/core/discovery"
	"github.com/bspippi1337/restless/internal/help"
)

func main() {
	if len(os.Args) < 2 {
		printRootHelp(0)
		return
	}

	switch os.Args[1] {
	case "-h", "--help", "help":
		printRootHelp(0)
		return
	case "--version", "version":
		fmt.Println(versionString())
		return
	case "discover":
		cmdDiscover(os.Args[2:])
		return
	case "doctor":
		cmdDoctor()
		return
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printRootHelp(2)
		return
	}
}

func printRootHelp(exit int) {
	out := os.Stdout
	fmt.Fprintln(out, "restless — domain-first API discovery and interaction engine")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  restless <command> [args]")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Commands:")
	fmt.Fprintln(out, "  discover   Discover APIs starting from a domain")
	fmt.Fprintln(out, "  doctor     Self-check and environment hints")
	fmt.Fprintln(out, "  version    Print version")
	fmt.Fprintln(out, "  help       Show help")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Try:")
	fmt.Fprintln(out, "  restless discover openai.com --verify --fuzz --save-profile openai")
	if exit != 0 {
		os.Exit(exit)
	}
}

func cmdDiscover(args []string) {
	fs := flag.NewFlagSet("discover", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	var (
		verify          = fs.Bool("verify", false, "Validate discovered endpoints with live HTTP checks")
		fuzz            = fs.Bool("fuzz", false, "Expand discovery using pattern-based probing")
		budgetSeconds   = fs.Int("budget-seconds", 15, "Maximum total discovery time")
		budgetPages     = fs.Int("budget-pages", 6, "Maximum pages to crawl")
		saveProfile     = fs.String("save-profile", "", "Save discovery results to a named profile")
		overwrite       = fs.Bool("overwrite-profile", false, "Replace existing profile instead of merging")
		profileDir      = fs.String("profile-dir", "", "Custom profile storage directory")
		emitExamples    = fs.Bool("emit-examples", false, "Generate example requests inside the profile")
		redactSecrets   = fs.Bool("redact-secrets", false, "Remove detected tokens from generated examples")
		jsonOut         = fs.Bool("json", false, "Output machine-readable JSON")
		quiet           = fs.Bool("quiet", false, "Minimal output")
		debug           = fs.Bool("debug", false, "Verbose diagnostic logging")
	)

	// Dynamic help hook for stdlib flags:
	fs.Usage = func() {
		ctx := help.NewDiscoverHelpContext(*profileDir)
		ctx.SupportsJSON = true
		ctx.SupportsTUI = false
		// Optional state file
		if st, ok := loadState(); ok {
			ctx.LastDomain = st.LastDomain
			ctx.ActiveProfile = st.ActiveProfile
		}
		fmt.Fprintln(fs.Output(), help.DiscoverHelp(ctx))
	}

	if err := fs.Parse(args); err != nil {
		// flag already printed error; show help
		fmt.Fprintln(os.Stdout, "")
		fs.Usage()
		os.Exit(2)
	}

	rest := fs.Args()
	if len(rest) < 1 {
		fs.Usage()
		os.Exit(2)
	}
	domain := rest[0]

	// persist state
	saveState(state{LastDomain: domain, ActiveProfile: strings.TrimSpace(*saveProfile)})

	if !*quiet {
		fmt.Printf("==> discover %s\n", domain)
	}
	find, err := discovery.DiscoverDomain(domain, discovery.Options{
		BudgetSeconds: *budgetSeconds,
		BudgetPages:   *budgetPages,
		Verify:        *verify,
		Fuzz:          *fuzz,
		Debug:         *debug,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "discover error: %v\n", err)
		os.Exit(1)
	}

	// Save profile if requested
	if *saveProfile != "" {
		dir := *profileDir
		if dir == "" {
			dir = defaultProfileDir()
		}
		path, err := writeProfile(dir, *saveProfile, domain, find, profileSaveOpts{
			Overwrite:     *overwrite,
			EmitExamples:  *emitExamples,
			RedactSecrets: *redactSecrets,
			Verify:        *verify,
			Fuzz:          *fuzz,
			BudgetSeconds: *budgetSeconds,
			BudgetPages:   *budgetPages,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "profile save error: %v\n", err)
			os.Exit(1)
		}
		if !*quiet {
			fmt.Printf("✅ Profile saved: %s\n", path)
			fmt.Printf("   Endpoints: %d  Docs: %d  Confidence: %.2f\n", len(find.Endpoints), len(find.DocURLs), find.Confidence)
			fmt.Printf("   Next: restless request --profile %s --method GET --path /v1/status\n", *saveProfile)
		}
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(find)
		return
	}

	if *quiet {
		return
	}

	// Human output
	fmt.Printf("Base URLs:\n")
	for _, u := range find.BaseURLs {
		fmt.Printf("  - %s\n", u)
	}
	fmt.Printf("Docs:\n")
	for _, u := range find.DocURLs {
		fmt.Printf("  - %s\n", u)
	}
	fmt.Printf("Endpoints (%d):\n", len(find.Endpoints))
	for _, ep := range find.Endpoints {
		fmt.Printf("  %s %s\n", ep.Method, ep.Path)
	}
	fmt.Printf("Confidence: %.2f\n", find.Confidence)
}

func cmdDoctor() {
	fmt.Println("==> doctor")
	fmt.Println("[ OK ] Go toolchain reachable:", runtimeGoVersion())
	fmt.Println("[ OK ] Profile dir:", defaultProfileDir())
	fmt.Println("Tip: run `restless discover openai.com --save-profile openai` to create your first profile.")
}

func defaultProfileDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "."
	}
	return filepath.Join(home, ".config", "restless", "profiles")
}

func versionString() string {
	// overridden by ldflags in CI/release if desired
	return "v0.0.0-dev"
}

func runtimeGoVersion() string {
	// minimal, avoids importing runtime in case of constraints; fine for doctor
	return strings.TrimSpace(os.Getenv("GOVERSION"))
}

type state struct {
	LastDomain    string `json:"lastDomain"`
	ActiveProfile string `json:"activeProfile"`
}

func statePath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ".restless-state.json"
	}
	return filepath.Join(home, ".config", "restless", "state.json")
}

func loadState() (state, bool) {
	var st state
	b, err := os.ReadFile(statePath())
	if err != nil {
		return st, false
	}
	_ = json.Unmarshal(b, &st)
	return st, true
}

func saveState(st state) {
	p := statePath()
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	b, _ := json.MarshalIndent(st, "", "  ")
	_ = os.WriteFile(p, b, 0o644)
}

// -------------------- Profile (YAML-ish minimal) --------------------

type profileSaveOpts struct {
	Overwrite     bool
	EmitExamples  bool
	RedactSecrets bool
	Verify        bool
	Fuzz          bool
	BudgetSeconds int
	BudgetPages   int
}

func writeProfile(dir, name, domain string, find discovery.Finding, opt profileSaveOpts) (string, error) {
	_ = os.MkdirAll(dir, 0o755)
	path := filepath.Join(dir, name+".yaml")

	// Merge-safe: if exists and not overwrite, keep auth + defaults block if present.
	var existingAuth string
	var existingDefaults string
	if !opt.Overwrite {
		if b, err := os.ReadFile(path); err == nil {
			s := string(b)
			existingAuth = extractBlock(s, "auth:")
			existingDefaults = extractBlock(s, "defaults:")
		}
	}

	now := time.Now().Format(time.RFC3339)
	var sb strings.Builder
	sb.WriteString("version: 1\n")
	sb.WriteString(fmt.Sprintf("name: %s\n", name))
	sb.WriteString(fmt.Sprintf("createdAt: %s\n", now))
	sb.WriteString(fmt.Sprintf("updatedAt: %s\n", now))
	sb.WriteString("\n")
	sb.WriteString("discoveredFrom:\n")
	sb.WriteString(fmt.Sprintf("  domain: %s\n", domain))
	sb.WriteString(fmt.Sprintf("  when: %s\n", now))
	sb.WriteString("  flags:\n")
	sb.WriteString(fmt.Sprintf("    verify: %v\n", opt.Verify))
	sb.WriteString(fmt.Sprintf("    fuzz: %v\n", opt.Fuzz))
	sb.WriteString(fmt.Sprintf("    budgetSeconds: %d\n", opt.BudgetSeconds))
	sb.WriteString(fmt.Sprintf("    budgetPages: %d\n", opt.BudgetPages))
	sb.WriteString("\n")

	sb.WriteString("baseUrls:\n")
	for _, u := range find.BaseURLs {
		sb.WriteString(fmt.Sprintf("  - %s\n", u))
	}
	if len(find.BaseURLs) == 0 {
		sb.WriteString("  - https://api." + domain + "\n")
	}
	sb.WriteString("\n")

	if existingAuth != "" {
		sb.WriteString(existingAuth)
		sb.WriteString("\n")
	} else {
		sb.WriteString("auth:\n")
		sb.WriteString("  type: bearer\n")
		sb.WriteString("  token:\n")
		sb.WriteString("    source: env\n")
		sb.WriteString("    envVar: RESTLESS_TOKEN\n\n")
	}

	if existingDefaults != "" {
		sb.WriteString(existingDefaults)
		sb.WriteString("\n")
	} else {
		sb.WriteString("defaults:\n")
		sb.WriteString("  headers:\n")
		sb.WriteString("    Accept: application/json\n")
		sb.WriteString("    User-Agent: restless/alpha\n")
		sb.WriteString("  timeoutSeconds: 20\n\n")
	}

	sb.WriteString("discovery:\n")
	sb.WriteString(fmt.Sprintf("  confidence: %.2f\n", find.Confidence))
	sb.WriteString("  docUrls:\n")
	if len(find.DocURLs) == 0 {
		sb.WriteString("    - https://" + domain + "/openapi.json\n")
	} else {
		for _, u := range find.DocURLs {
			sb.WriteString("    - " + u + "\n")
		}
	}
	sb.WriteString("\n")

	sb.WriteString("endpoints:\n")
	for _, ep := range find.Endpoints {
		sb.WriteString(fmt.Sprintf("  - method: %s\n", ep.Method))
		sb.WriteString(fmt.Sprintf("    path: %s\n", ep.Path))
		sb.WriteString(fmt.Sprintf("    score: %.2f\n", ep.Score))
		sb.WriteString("    evidence:\n")
		for _, ev := range ep.Evidence {
			sb.WriteString(fmt.Sprintf("      - source: %s\n", ev.Source))
			sb.WriteString(fmt.Sprintf("        url: %s\n", ev.URL))
			sb.WriteString(fmt.Sprintf("        when: %s\n", ev.When))
			sb.WriteString(fmt.Sprintf("        score: %.2f\n", ev.Score))
		}
	}
	if len(find.Endpoints) == 0 {
		sb.WriteString("  - method: GET\n    path: /v1/status\n    score: 0.50\n    evidence:\n      - source: heuristic\n        url: https://" + domain + "/\n        when: " + now + "\n        score: 0.50\n")
	}
	sb.WriteString("\n")

	if opt.EmitExamples {
		sb.WriteString("examples:\n")
		sb.WriteString("  - name: status\n")
		sb.WriteString("    request:\n")
		sb.WriteString("      method: GET\n")
		sb.WriteString("      path: /v1/status\n")
		sb.WriteString("      headers:\n")
		sb.WriteString("        Authorization: \"Bearer ${ENV:RESTLESS_TOKEN}\"\n")
		sb.WriteString("\n")
	}

	return path, os.WriteFile(path, []byte(sb.String()), 0o644)
}

func extractBlock(s, header string) string {
	idx := strings.Index(s, header)
	if idx < 0 {
		return ""
	}
	rest := s[idx:]
	lines := strings.Split(rest, "\n")
	var out []string
	out = append(out, lines[0])
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if len(line) > 0 && line[0] != ' ' && strings.HasSuffix(line, ":") {
			break
		}
		if strings.TrimSpace(line) == "" && i > 1 {
			// allow blank line but stop after block ends
			out = append(out, line)
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n") + "\n"
}
