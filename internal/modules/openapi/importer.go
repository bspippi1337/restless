package openapi

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Import fetches from URL or reads from file, then stores raw content in ~/.restless/openapi/<id>.yaml|json
// v1 skeleton: we don't parse endpoints yet; we just cache and create an index entry.
func Import(source string) (SpecIndex, error) {
	if source == "" {
		return SpecIndex{}, errors.New("empty source")
	}

	id := idFor(source)
	dir, err := cacheDir()
	if err != nil {
		return SpecIndex{}, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return SpecIndex{}, err
	}

	rawPath := filepath.Join(dir, id+".raw")
	var data []byte

	if looksLikeURL(source) {
		resp, err := http.Get(source) //nolint:gosec
		if err != nil {
			return SpecIndex{}, err
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return SpecIndex{}, err
		}
		data = b
	} else {
		b, err := os.ReadFile(source)
		if err != nil {
			return SpecIndex{}, err
		}
		data = b
	}

	if err := os.WriteFile(rawPath, data, 0o644); err != nil {
		return SpecIndex{}, err
	}

	idx := SpecIndex{
		ID:       id,
		Source:   source,
		Imported: time.Now().Unix(),
		Title:    "",
		Version:  "",
		RawPath:  rawPath,
	}

	if err := SaveIndex(idx); err != nil {
		return SpecIndex{}, err
	}
	return idx, nil
}

func looksLikeURL(s string) bool {
	return len(s) > 7 && (s[:7] == "http://" || (len(s) > 8 && s[:8] == "https://"))
}
