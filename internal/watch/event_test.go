package watch

import (
	"testing"
	"time"

	"github.com/bspippi1337/restless/internal/events"
)

func shouldEmit(last map[string]time.Time, path string, now time.Time, debounce time.Duration) bool {
	if t, ok := last[path]; ok && now.Sub(t) < debounce {
		return false
	}
	last[path] = now
	return true
}

func TestDebounceAllowsFirstEvent(t *testing.T) {
	last := map[string]time.Time{}
	now := time.Now()

	if !shouldEmit(last, "demo.txt", now, 250*time.Millisecond) {
		t.Fatalf("first event should be emitted")
	}
}

func TestDebounceSuppressesRapidRepeat(t *testing.T) {
	last := map[string]time.Time{}
	now := time.Now()

	if !shouldEmit(last, "demo.txt", now, 250*time.Millisecond) {
		t.Fatalf("first event should be emitted")
	}

	if shouldEmit(last, "demo.txt", now.Add(10*time.Millisecond), 250*time.Millisecond) {
		t.Fatalf("rapid repeat should be suppressed")
	}
}

func TestDebounceAllowsLaterRepeat(t *testing.T) {
	last := map[string]time.Time{}
	now := time.Now()

	if !shouldEmit(last, "demo.txt", now, 250*time.Millisecond) {
		t.Fatalf("first event should be emitted")
	}

	if !shouldEmit(last, "demo.txt", now.Add(500*time.Millisecond), 250*time.Millisecond) {
		t.Fatalf("later repeat should be emitted")
	}
}

func TestEventShapeForWatcher(t *testing.T) {
	ev := events.New("fsnotify", "filesystem", "demo.txt")
	ev.Op = "WRITE"

	if ev.Source != "fsnotify" || ev.Kind != "filesystem" || ev.Op != "WRITE" {
		t.Fatalf("unexpected watcher event: %#v", ev)
	}
}
