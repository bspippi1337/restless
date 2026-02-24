package openapi

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

type SpecIndex struct {
	ID        string `json:"id"`
	Source    string `json:"source"` // URL or file path
	Imported  int64  `json:"imported_unix"`
	Title     string `json:"title"`
	Version   string `json:"version"`
	BaseURL   string `json:"base_url"`
	RawPath   string `json:"raw_path"`
	IndexPath string `json:"index_path"`
}

func cacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".restless", "openapi"), nil
}

func idFor(source string) string {
	h := sha1.Sum([]byte(source))
	return hex.EncodeToString(h[:])
}

func SaveIndex(idx SpecIndex) error {
	dir, err := cacheDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	p := filepath.Join(dir, idx.ID+".json")
	idx.IndexPath = p
	b, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o644)
}

func LoadIndex(id string) (SpecIndex, error) {
	dir, err := cacheDir()
	if err != nil {
		return SpecIndex{}, err
	}
	p := filepath.Join(dir, id+".json")
	b, err := os.ReadFile(p)
	if err != nil {
		return SpecIndex{}, err
	}
	var idx SpecIndex
	if err := json.Unmarshal(b, &idx); err != nil {
		return SpecIndex{}, err
	}
	idx.IndexPath = p
	return idx, nil
}

func ListIndexFiles() ([]string, error) {
	dir, err := cacheDir()
	if err != nil {
		return nil, err
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) == ".json" {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	return out, nil
}
