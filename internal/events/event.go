package events

import "time"

// Event is the normalized signal that flows through Restless runtime layers.
// Filesystem watchers, future HTTP probes, and other sensors should emit this shape
// so downstream pipeline and observability code can stay small and composable.
type Event struct {
	ID        string            `json:"id"`
	Source    string            `json:"source"`
	Kind      string            `json:"kind"`
	Path      string            `json:"path"`
	Op        string            `json:"op,omitempty"`
	Time      time.Time         `json:"time"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

func New(source, kind, path string) Event {
	now := time.Now().UTC()
	return Event{
		ID:       now.Format("20060102T150405.000000000Z07:00") + ":" + source + ":" + kind + ":" + path,
		Source:   source,
		Kind:     kind,
		Path:     path,
		Time:     now,
		Metadata: map[string]string{},
	}
}
