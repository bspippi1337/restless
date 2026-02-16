// internal/help/discover.go
package help

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// HelpContext is intentionally small and dependency-free.
// Populate what you can; leave the rest empty.
type HelpContext struct {
	TerminalWidth int

	// Optional: makes help feel "alive"
	LastDomain    string
	ActiveProfile string

	// Profiles
	ProfileDir string
	Profiles   []string

	// Feature signals (if your build/runtime supports it)
	SupportsJSON bool
	SupportsTUI  bool
}

// DiscoverHelp returns a dynamic help text for: `restless discover --help`.
func DiscoverHelp(ctx HelpContext) string {
	w := ctx.TerminalWidth
	if w <= 0 {
		w = detectWidth(92)
	}
	if ctx.ProfileDir == "" {
		ctx.ProfileDir = defaultProfileDir()
	}
	if len(ctx.Profiles) == 0 {
		ctx.Profiles = listProfileNames(ctx.ProfileDir)
	}
	sort.Strings(ctx.Profiles)

	var b strings.Builder

	title(&b, "restless discover", "domain-first API discovery engine")
	blank(&b)

	para(&b, w, "Usage:", "restless discover <domain> [flags]")
	blank(&b)

	para(&b, w, "Description:",
		"Discover APIs starting from a domain. Restless scans documentation, OpenAPI specs, sitemaps, common paths, and optionally fuzzes for likely endpoints.")
	blank(&b)

	// Context-driven “you are here” hints
	if ctx.LastDomain != "" {
		callout(&b, w, "Last domain", ctx.LastDomain)
	}
	if ctx.ActiveProfile != "" {
		callout(&b, w, "Active profile", ctx.ActiveProfile)
	}
	if len(ctx.Profiles) == 0 {
		callout(&b, w, "Profiles", "none saved yet")
	} else {
		callout(&b, w, "Profiles", fmt.Sprintf("%d saved", len(ctx.Profiles)))
	}
	blank(&b)

	section(&b, "Quickstart (recommended)")
	if len(ctx.Profiles) == 0 {
		cmd(&b, "restless discover openai.com --verify --fuzz --budget-seconds 20 --budget-pages 8 --save-profile openai")
		para(&b, w, "Why profiles?",
			"Profiles turn discovery into a reusable configuration for requests, TUI auto-fill, and repeatable automation.")
	} else {
		name := ctx.Profiles[0]
		cmd(&b, fmt.Sprintf("restless discover openai.com --verify --fuzz --save-profile %s", shellSafe(name)))
		para(&b, w, "Tip:",
			"Use --save-profile to refresh an existing profile (merge-safe).")
	}
	blank(&b)

	section(&b, "Examples")
	cmd(&b, "restless discover openai.com")
	cmd(&b, "restless discover openai.com --verify --fuzz")
	cmd(&b, "restless discover openai.com --budget-seconds 20 --budget-pages 8")
	cmd(&b, "restless discover openai.com --save-profile openai")
	cmd(&b, "restless discover openai.com --save-profile openai --overwrite-profile")
	cmd(&b, "restless discover openai.com --save-profile openai --profile-dir ./profiles")
	blank(&b)

	if len(ctx.Profiles) > 0 {
		section(&b, "Saved profiles")
		list(&b, w, ctx.Profiles, 8)
		blank(&b)
		para(&b, w, "Profile location:", ctx.ProfileDir)
		blank(&b)
	} else {
		para(&b, w, "Profile location:", ctx.ProfileDir)
		blank(&b)
	}

	section(&b, "Flags")
	flag(&b, "--verify", "Validate discovered endpoints with live HTTP checks.")
	flag(&b, "--fuzz", "Expand discovery using pattern-based probing (doc-guided when docs are found).")
	flag(&b, "--budget-seconds <int>", "Maximum total discovery time. (default 15)")
	flag(&b, "--budget-pages <int>", "Maximum pages to crawl. (default 6)")
	flag(&b, "--save-profile <name>", "Save discovery results to a named profile.")
	flag(&b, "--overwrite-profile", "Replace existing profile instead of merging. (dangerous)")
	flag(&b, "--profile-dir <path>", "Custom profile storage directory.")
	flag(&b, "--emit-examples", "Generate example requests inside the profile.")
	flag(&b, "--redact-secrets", "Remove detected tokens from generated examples.")
	if ctx.SupportsJSON {
		flag(&b, "--json", "Output machine-readable JSON.")
	} else {
		flag(&b, "--json", "Output machine-readable JSON. (if supported in your build)")
	}
	flag(&b, "--quiet", "Minimal output.")
	flag(&b, "--debug", "Verbose diagnostic logging.")
	blank(&b)

	section(&b, "Output")
	para(&b, w, "",
		"By default discover prints domain, base URLs, endpoints (method + path), confidence score, and evidence sources.")
	para(&b, w, "",
		"When --save-profile is used, discover writes a profile file and prints the path plus counts.")
	blank(&b)

	section(&b, "Exit codes")
	lines(&b,
		"0  Success",
		"1  Error",
		"2  Completed but no endpoints found",
	)
	blank(&b)

	section(&b, "Notes")
	lines(&b,
		"• Discovery is read-only.",
		"• Fuzz mode never performs destructive requests.",
		"• Profiles should reference secrets via environment variables (not stored plaintext).",
	)
	blank(&b)

	return trimEnd(b.String())
}

// NewDiscoverHelpContext is a convenience. Call it from your discover command.
func NewDiscoverHelpContext(profileDir string) HelpContext {
	return HelpContext{
		TerminalWidth: detectWidth(92),
		ProfileDir:    profileDir,
		Profiles:      listProfileNames(profileDir),
		SupportsJSON:  true,
		SupportsTUI:   true,
	}
}

// --------- Tiny renderer (standalone; no other files needed) ---------

func detectWidth(fallback int) int {
	fd := int(os.Stdout.Fd())
	if term.IsTerminal(fd) {
		if w, _, err := term.GetSize(fd); err == nil && w > 0 {
			if w < 60 {
				return 60
			}
			if w > 120 {
				return 120
			}
			return w
		}
	}
	return fallback
}

func defaultProfileDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "."
	}
	return filepath.Join(home, ".config", "restless", "profiles")
}

func listProfileNames(dir string) []string {
	var out []string
	ents, err := os.ReadDir(dir)
	if err != nil {
		return out
	}
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".yaml") {
			out = append(out, strings.TrimSuffix(name, ".yaml"))
		} else if strings.HasSuffix(name, ".yml") {
			out = append(out, strings.TrimSuffix(name, ".yml"))
		}
	}
	return out
}

func title(b *strings.Builder, a, sub string) {
	b.WriteString(a)
	if sub != "" {
		b.WriteString(" — ")
		b.WriteString(sub)
	}
	b.WriteString("\n")
}

func section(b *strings.Builder, s string) {
	b.WriteString(s)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("-", min(72, max(12, len(s)))))
	b.WriteString("\n")
}

func cmd(b *strings.Builder, s string) {
	b.WriteString("  ")
	b.WriteString(s)
	b.WriteString("\n")
}

func flag(b *strings.Builder, name, desc string) {
	b.WriteString("  ")
	b.WriteString(padRight(name, 22))
	b.WriteString(desc)
	b.WriteString("\n")
}

func callout(b *strings.Builder, width int, label, value string) {
	line := fmt.Sprintf("%s: %s", label, value)
	b.WriteString(wrapLine(line, width))
	b.WriteString("\n")
}

func lines(b *strings.Builder, ls ...string) {
	for _, s := range ls {
		b.WriteString("  ")
		b.WriteString(s)
		b.WriteString("\n")
	}
}

func list(b *strings.Builder, width int, items []string, maxItems int) {
	if len(items) == 0 {
		lines(b, "  (none)")
		return
	}
	if maxItems > 0 && len(items) > maxItems {
		items = append(items[:maxItems], "…")
	}
	for _, it := range items {
		b.WriteString("  • ")
		b.WriteString(wrapLine(it, width-4))
		b.WriteString("\n")
	}
}

func para(b *strings.Builder, width int, head, body string) {
	if head != "" {
		b.WriteString(head)
		b.WriteString("\n")
	}
	if body == "" {
		return
	}
	for _, line := range wrap(body, width) {
		b.WriteString("  ")
		b.WriteString(line)
		b.WriteString("\n")
	}
}

func blank(b *strings.Builder) { b.WriteString("\n") }

func wrap(s string, width int) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}
	}
	words := strings.Fields(s)
	var lines []string
	var cur strings.Builder
	for _, w := range words {
		if cur.Len() == 0 {
			cur.WriteString(w)
			continue
		}
		if cur.Len()+1+len(w) > width-2 {
			lines = append(lines, cur.String())
			cur.Reset()
			cur.WriteString(w)
		} else {
			cur.WriteString(" ")
			cur.WriteString(w)
		}
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	return lines
}

func wrapLine(s string, width int) string {
	ls := wrap(s, width)
	if len(ls) == 0 {
		return ""
	}
	if len(ls) == 1 {
		return ls[0]
	}
	return strings.Join(ls, "\n  ")
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s + " "
	}
	return s + strings.Repeat(" ", n-len(s))
}

func trimEnd(s string) string { return strings.TrimRight(s, "\n") }

func shellSafe(s string) string {
	if strings.ContainsAny(s, " \t") {
		return strconv.Quote(s)
	}
	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
