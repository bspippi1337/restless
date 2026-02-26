package docparse

import (
	"sort"
	"strings"
)

// EndpointLite is intentionally small to avoid an import cycle.
// discovery owns the richer Endpoint type (FullURL, Evidences, etc).
type EndpointLite struct {
	Method string
	Path   string
}

func EndpointsFromOpenAPI(o *OpenAPI) []EndpointLite {
	out := []EndpointLite{}
	if o == nil {
		return out
	}

	for path, methods := range o.Paths {
		for m := range methods {
			mm := strings.ToUpper(strings.TrimSpace(m))
			if mm == "" {
				mm = "GET"
			}
			out = append(out, EndpointLite{Method: mm, Path: path})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Path == out[j].Path {
			return out[i].Method < out[j].Method
		}
		return out[i].Path < out[j].Path
	})

	return out
}
