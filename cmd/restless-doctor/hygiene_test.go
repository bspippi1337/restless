package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootAuditAllowsConventionalFiles(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"README.md", "SECURITY.md", "CHANGELOG.md", "CONTRIBUTING.md", "CODE_OF_CONDUCT.md", "install.sh", "LICENSE"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("test"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(root, "docs"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := auditRoot(root); err != nil {
		t.Fatal(err)
	}
}

func TestRootAuditRejectsHistoricalFiles(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"MIGRATION_v2.md", "full_verify.sh", "SHA"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("legacy"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	err := auditRoot(root)
	if err == nil || !strings.Contains(err.Error(), "MIGRATION_v2.md") || !strings.Contains(err.Error(), "full_verify.sh") {
		t.Fatalf("expected a root violation, got %v", err)
	}
	for _, name := range []string{"MIGRATION_v2.md", "full_verify.sh", "SHA"} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Fatalf("audit modified %s: %v", name, err)
		}
	}
}

func TestCommandsAreShellFree(t *testing.T) {
	checks := commands("restless")
	if len(checks) != 5 {
		t.Fatalf("expected 5 checks, got %d", len(checks))
	}
	for _, step := range checks {
		if step.tool != "go" && step.tool != "gofmt" {
			t.Fatalf("unexpected executable %q", step.tool)
		}
	}
}
