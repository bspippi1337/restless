package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Entry struct {
	When       string `json:"when"`
	Profile    string `json:"profile"`
	Name       string `json:"name,omitempty"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	BaseURL    string `json:"baseUrl"`
	StatusCode int    `json:"statusCode"`
	LatencyMs  int64  `json:"latencyMs"`
	OK         bool   `json:"ok"`
}

func DefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "."
	}
	return filepath.Join(home, ".config", "restless", "history")
}

func Append(profile string, e Entry) error {
	dir := DefaultDir()
	_ = os.MkdirAll(dir, 0o755)
	path := filepath.Join(dir, profile+".jsonl")
	if e.When == "" {
		e.When = time.Now().Format(time.RFC3339)
	}
	b, _ := json.Marshal(e)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(b, '\n'))
	return err
}
