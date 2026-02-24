package gui

import "sync"

type AtomString struct {
	mu sync.RWMutex
	v  string
}

func (a *AtomString) Get() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.v
}
func (a *AtomString) Set(v string) {
	a.mu.Lock()
	a.v = v
	a.mu.Unlock()
}

type State struct {
	URL         AtomString
	Method      AtomString
	HeadersJSON AtomString
	Status      AtomString
	Body        AtomString
	Error       AtomString
	SmartHints  []string
	RawLinks    []string
}

func NewState() *State {
	s := &State{}
	s.Status.Set("HTTP -")
	return s
}
