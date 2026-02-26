package cli

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type Mode string

const (
	ModeHeuristic Mode = "heuristic"
	ModeSpec      Mode = "spec"
)

type Endpoint struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Source string `json:"source,omitempty"` // spec|heuristic
}

type Session struct {
	Name         string     `json:"name"`
	BaseURL      string     `json:"base_url"`
	Mode         Mode       `json:"mode"`
	Endpoints    []Endpoint `json:"endpoints,omitempty"`
	LastCall     string     `json:"last_call,omitempty"`
	RequestCount int        `json:"request_count,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type State struct {
	SessionName string
	NoHeader    bool

	Session Session
}

func NewState() *State {
	return &State{
		SessionName: "default",
		Session: Session{
			Name: "default",
			Mode: ModeHeuristic,
		},
	}
}

func (s *State) Load() error {
	// Refresh name because cobra flag is already bound.
	s.Session.Name = s.SessionName

	p, err := sessionPath(s.SessionName)
	if err != nil {
		return err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	var sess Session
	if err := json.Unmarshal(b, &sess); err != nil {
		return err
	}
	s.Session = sess
	// Ensure current flag name wins.
	s.Session.Name = s.SessionName
	return nil
}

func (s *State) Save() error {
	s.Session.Name = s.SessionName
	s.Session.UpdatedAt = time.Now()

	p, err := sessionPath(s.SessionName)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(s.Session, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, append(b, '\n'), 0o644)
}

func sessionPath(name string) (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "restless", "sessions", name+".json"), nil
}
