package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Route struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type ScanResult struct {
	BaseURL   string  `json:"base_url"`
	Endpoints []Route `json:"endpoints"`
}

type State struct {
	LastScan ScanResult `json:"last_scan"`
}

func path() string {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return filepath.Join(xdg, "restless", "state.json")
	}

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ".restless_state.json"
	}

	return filepath.Join(home, ".local", "state", "restless", "state.json")
}

func legacyPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}

	return filepath.Join(home, ".restless_state.json")
}

func ensureDir(p string) {
	dir := filepath.Dir(p)
	_ = os.MkdirAll(dir, 0700)
}

func HasScan(s State) bool {
	return s.LastScan.BaseURL != "" || len(s.LastScan.Endpoints) > 0
}

func Load() (State, string, error) {
	p := path()

	b, err := os.ReadFile(p)
	if err != nil {
		legacy := legacyPath()

		if legacy != "" {
			if lb, lerr := os.ReadFile(legacy); lerr == nil {
				var s State
				if json.Unmarshal(lb, &s) == nil {
					return s, legacy, nil
				}
			}
		}

		if os.IsNotExist(err) {
			return State{}, p, nil
		}

		return State{}, p, err
	}

	var s State
	if err := json.Unmarshal(b, &s); err != nil {
		return State{}, p, err
	}

	return s, p, nil
}

func Save(s State) (string, error) {
	p := path()
	ensureDir(p)

	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return p, err
	}

	return p, os.WriteFile(p, b, 0600)
}
