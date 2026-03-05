package export

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/bspippi1337/restless/internal/core/types"
)

func SaveJSONArtifact(name string, resp types.Response) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	ts := time.Now().Format("20060102-150405")
	dir := filepath.Join(home, ".restless", "artifacts", ts)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if name == "" {
		name = "response"
	}
	p := filepath.Join(dir, name+".json")
	b, err := json.MarshalIndent(map[string]any{
		"status":   resp.StatusCode,
		"headers":  resp.Headers,
		"duration": resp.DurationMs,
		"body":     string(resp.Body),
	}, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(p, b, 0o644); err != nil {
		return "", err
	}
	return p, nil
}
