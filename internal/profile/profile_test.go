package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_EmptyName(t *testing.T) {
	if _, err := Load(t.TempDir(), ""); err == nil {
		t.Fatalf("expected error")
	}
}

func TestLoad_BasicYAML(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "openai.yaml")
	content := `baseUrls:
  - https://api.example.com
auth:
  type: bearer
  token:
    env: OPENAI_API_KEY
defaults:
  timeoutSeconds: 12
  headers:
    X-App: restless
endpoints:
  - method: GET
    path: /v1/status
    score: 0.9
`
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	pr, err := Load(dir, "openai")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if pr.Name != "openai" {
		t.Fatalf("name: %q", pr.Name)
	}
	if len(pr.BaseURLs) != 1 || pr.BaseURLs[0] != "https://api.example.com" {
		t.Fatalf("base urls: %#v", pr.BaseURLs)
	}
	if pr.AuthType != "bearer" || pr.AuthEnv != "OPENAI_API_KEY" {
		t.Fatalf("auth: type=%q env=%q", pr.AuthType, pr.AuthEnv)
	}
	if pr.TimeoutS != 12 {
		t.Fatalf("timeout: %d", pr.TimeoutS)
	}
	if pr.Defaults["X-App"] != "restless" {
		t.Fatalf("headers/defaults not parsed: %#v", pr.Defaults)
	}
	if len(pr.Endpoints) != 1 || pr.Endpoints[0].Path != "/v1/status" {
		t.Fatalf("endpoints: %#v", pr.Endpoints)
	}
}
