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
	RawPath   string `json:"raw_path"` // stored raw file
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
	return idx, nil
}
