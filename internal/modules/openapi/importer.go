package openapi

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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

	raw, ext, err := readSource(source)
	if err != nil {
		return SpecIndex{}, err
	}

	rawPath := filepath.Join(dir, id+ext)
	if err := os.WriteFile(rawPath, raw, 0o644); err != nil {
		return SpecIndex{}, err
	}

	// Parse minimal metadata if possible
	title := ""
	ver := ""
	base := ""
	if spec, err := LoadSpec(raw); err == nil {
		title = spec.Info.Title
		ver = spec.Info.Version
		base = spec.BaseURL()
	}

	idx := SpecIndex{
		ID:       id,
		Source:   source,
		Imported: time.Now().Unix(),
		Title:    title,
		Version:  ver,
		BaseURL:  base,
		RawPath:  rawPath,
	}
	if err := SaveIndex(idx); err != nil {
		return SpecIndex{}, err
	}
	return idx, nil
}

func readSource(source string) (raw []byte, ext string, err error) {
	if looksLikeURL(source) {
		resp, err := http.Get(source) //nolint:gosec
		if err != nil {
			return nil, "", err
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", err
		}
		ext = sniffExt(source, resp.Header.Get("Content-Type"), b)
		return b, ext, nil
	}

	b, err := os.ReadFile(source)
	if err != nil {
		return nil, "", err
	}
	ext = sniffExt(source, "", b)
	return b, ext, nil
}

func looksLikeURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func sniffExt(source, contentType string, raw []byte) string {
	// If filename gives clue
	ls := strings.ToLower(source)
	if strings.HasSuffix(ls, ".json") {
		return ".json"
	}
	if strings.HasSuffix(ls, ".yaml") || strings.HasSuffix(ls, ".yml") {
		return ".yaml"
	}
	ct := strings.ToLower(contentType)
	if strings.Contains(ct, "json") {
		return ".json"
	}
	if strings.Contains(ct, "yaml") || strings.Contains(ct, "yml") {
		return ".yaml"
	}
	// sniff content
	trim := strings.TrimSpace(string(raw))
	if strings.HasPrefix(trim, "{") || strings.HasPrefix(trim, "[") {
		return ".json"
	}
	return ".yaml"
}
