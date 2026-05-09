package events

import "testing"

func TestNewEvent(t *testing.T) {
	ev := New("fsnotify", "filesystem", "demo.txt")

	if ev.Source != "fsnotify" {
		t.Fatalf("unexpected source %q", ev.Source)
	}

	if ev.Kind != "filesystem" {
		t.Fatalf("unexpected kind %q", ev.Kind)
	}

	if ev.Path != "demo.txt" {
		t.Fatalf("unexpected path %q", ev.Path)
	}

	if ev.ID == "" {
		t.Fatalf("expected generated ID")
	}

	if ev.Metadata == nil {
		t.Fatalf("metadata map must be initialized")
	}
}
