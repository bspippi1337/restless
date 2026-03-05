package council

import "sync"

type Blackboard struct {
	mu       sync.RWMutex
	Findings []Finding
}

func NewBlackboard() *Blackboard {
	return &Blackboard{}
}

func (b *Blackboard) Publish(f Finding) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Findings = append(b.Findings, f)
}

func (b *Blackboard) List() []Finding {
	b.mu.RLock()
	defer b.mu.RUnlock()

	out := make([]Finding, len(b.Findings))
	copy(out, b.Findings)
	return out
}
