package snippets

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveLoadListTouch(t *testing.T) {
	dir := t.TempDir()
	sn := Snippet{
		Name:      "List Models",
		Profile:   "openai",
		Method:    "GET",
		Path:      "/v1/models",
		Headers:   map[string]string{"Accept": "application/json"},
		Notes:     "Smoke test",
		Tags:      []string{"core", "read"},
		CreatedAt: time.Now().Add(-time.Hour).Format(time.RFC3339),
	}

	savedPath, err := Save(dir, sn, true)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if _, err := os.Stat(savedPath); err != nil {
		t.Fatalf("expected file: %v", err)
	}
	if filepath.Ext(savedPath) != ".yaml" {
		t.Fatalf("expected yaml file, got %q", savedPath)
	}

	got, err := Load(dir, "openai", sn.Name)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Name != sn.Name || got.Path != sn.Path || got.Method != sn.Method {
		t.Fatalf("roundtrip mismatch: %#v", got)
	}

	if err := Touch(dir, got); err != nil {
		t.Fatalf("touch: %v", err)
	}

	lst, err := List(dir, "openai")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(lst) != 1 {
		t.Fatalf("expected 1 snippet, got %d", len(lst))
	}
	if lst[0].UseCount < 1 {
		t.Fatalf("expected UseCount >= 1, got %d", lst[0].UseCount)
	}
}

func TestSave_RequiresNameAndProfile(t *testing.T) {
	_, err := Save(t.TempDir(), Snippet{Name: "", Profile: "p"}, true)
	if err == nil {
		t.Fatalf("expected error")
	}
	_, err = Save(t.TempDir(), Snippet{Name: "n", Profile: ""}, true)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestTouchWithResult(t *testing.T) {
	dir := t.TempDir()
	sn := Snippet{Name: "Ping", Profile: "openai", Method: "GET", Path: "/v1/ping", CreatedAt: time.Now().Format(time.RFC3339)}
	_, err := Save(dir, sn, true)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	sn2, err := Load(dir, "openai", "Ping")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if err := TouchWithResult(dir, sn2, true, 123); err != nil {
		t.Fatalf("touchwithresult: %v", err)
	}
	sn3, _ := Load(dir, "openai", "Ping")
	if sn3.Successes < 1 || sn3.AvgLatencyMs <= 0 {
		t.Fatalf("expected stats updated: %#v", sn3)
	}
}
