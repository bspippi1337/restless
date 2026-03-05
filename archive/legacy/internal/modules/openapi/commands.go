package openapi

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func ListSpecs() error {
	files, err := ListIndexFiles()
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, p := range files {
		id := strings.TrimSuffix(filepath.Base(p), ".json")
		idx, err := LoadIndex(id)
		if err != nil {
			// best effort
			fmt.Printf("%s  (failed to read index)\n", id)
			continue
		}
		title := idx.Title
		if title == "" {
			title = "(no title)"
		}
		ver := idx.Version
		if ver == "" {
			ver = "(no version)"
		}
		base := idx.BaseURL
		if base == "" {
			base = "(no base url)"
		}
		fmt.Printf("%s  %s  %s  base=%s  src=%s\n", idx.ID, title, ver, base, idx.Source)
	}
	return nil
}

type Endpoint struct {
	Method      string
	Path        string
	Summary     string
	OperationID string
}

func ListEndpoints(id string) ([]Endpoint, error) {
	idx, err := LoadIndex(id)
	if err != nil {
		return nil, err
	}
	spec, err := LoadSpecFromFile(idx.RawPath)
	if err != nil {
		return nil, err
	}

	var out []Endpoint
	for path, item := range spec.Paths {
		for method, op := range item {
			out = append(out, Endpoint{
				Method:      strings.ToUpper(method),
				Path:        path,
				Summary:     op.Summary,
				OperationID: op.OperationID,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Path == out[j].Path {
			return out[i].Method < out[j].Method
		}
		return out[i].Path < out[j].Path
	})
	return out, nil
}

func PrintEndpoints(id string) error {
	eps, err := ListEndpoints(id)
	if err != nil {
		return err
	}
	for _, e := range eps {
		s := e.Summary
		if s == "" {
			s = "-"
		}
		op := e.OperationID
		if op == "" {
			op = "-"
		}
		fmt.Printf("%-6s %-40s  %s  (op:%s)\n", e.Method, e.Path, s, op)
	}
	return nil
}
