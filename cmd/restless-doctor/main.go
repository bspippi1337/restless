// RESTLESS doctor is a shell-free, cross-platform replacement for scripts/doctor.sh.
// It verifies the module, checks formatting, runs go vet/test, and builds to a temp dir.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"
)

type check struct {
	title string
	tool  string
	args  []string
}

func commands(output string) []check {
	return []check{
		{"verify modules", "go", []string{"mod", "verify"}},
		{"formatting", "gofmt", []string{"-l", "."}},
		{"static analysis", "go", []string{"vet", "./..."}},
		{"tests", "go", []string{"test", "./..."}},
		{"build", "go", []string{"build", "-trimpath", "-o", output, "./cmd/restless"}},
	}
}

func main() {
	repo := flag.String("repo", ".", "path to the Restless repository")
	timeout := flag.Duration("timeout", 10*time.Minute, "maximum total verification time")
	auditOnly := flag.Bool("audit-only", false, "verify root layout without running Go commands")
	flag.Parse()
	if flag.NArg() != 0 || *timeout <= 0 {
		fmt.Fprintln(os.Stderr, "usage: restless-doctor [--repo PATH] [--timeout 10m] [--audit-only]")
		os.Exit(2)
	}
	root, err := filepath.Abs(*repo)
	if err != nil {
		fatal(err)
	}
	for _, p := range []string{"go.mod", filepath.Join("cmd", "restless")} {
		if _, err := os.Stat(filepath.Join(root, p)); err != nil {
			fatal(fmt.Errorf("not a Restless checkout (missing %s): %w", p, err))
		}
	}
	if err := auditRoot(root); err != nil {
		fatal(err)
	}
	if *auditOnly {
		fmt.Println("RESTLESS doctor: root layout PASS")
		return
	}
	temp, err := os.MkdirTemp("", "restless-doctor-*")
	if err != nil {
		fatal(err)
	}
	defer os.RemoveAll(temp)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	ctx, stop := context.WithTimeout(ctx, *timeout)
	defer stop()
	output := filepath.Join(temp, "restless")
	if os.PathSeparator == '\\' {
		output += ".exe"
	}
	fmt.Printf("RESTLESS doctor: verifying %s (no source changes)\n", root)
	for _, step := range commands(output) {
		fmt.Printf("==> %s\n", step.title)
		start := time.Now()
		cmd := exec.CommandContext(ctx, step.tool, step.args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		if len(out) > 0 {
			fmt.Print(string(out))
		}
		if err != nil {
			fatal(fmt.Errorf("%s failed after %s: %w", step.title, time.Since(start).Round(time.Millisecond), err))
		}
		if step.tool == "gofmt" && strings.TrimSpace(string(out)) != "" {
			fatal(errors.New("unformatted Go files found; run gofmt on those files before merging"))
		}
		fmt.Printf("OK (%s)\n", time.Since(start).Round(time.Millisecond))
	}
	fmt.Println("RESTLESS doctor: PASS")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "ERROR:", err)
	os.Exit(1)
}
