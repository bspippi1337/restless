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
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".restless_state.json")
}

func Load() (State, string, error) {

	p := path()

	b, err := os.ReadFile(p)
	if err != nil {
		return State{}, p, nil
	}

	var s State
	json.Unmarshal(b, &s)

	return s, p, nil
}

func Save(s State) (string, error) {

	p := path()

	b, _ := json.MarshalIndent(s, "", "  ")

	os.WriteFile(p, b, 0644)

	return p, nil
}
