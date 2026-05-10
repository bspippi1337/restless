package events

import (
	"encoding/json"
	"testing"
)

func TestEventJSONSnapshot(t *testing.T) {
	ev := New("fsnotify", "filesystem", "demo.txt")
	ev.Op = "WRITE"
	ev.Metadata["engine"] = "watch"

	b, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	got := string(b)

	checks := []string{
		`"source":"fsnotify"`,
		`"kind":"filesystem"`,
		`"path":"demo.txt"`,
		`"op":"WRITE"`,
	}

	for _, c := range checks {
		if !contains(got, c) {
			t.Fatalf("snapshot missing %q in %s", c, got)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}()
}
