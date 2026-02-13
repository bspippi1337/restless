package docparse

import (
	"sort"
	"strings"

	"github.com/bspippi1337/restless/internal/core/model"
)

func EndpointsFromOpenAPI(o *OpenAPI) []model.Endpoint {
	out := []model.Endpoint{}
	for path, methods := range o.Paths {
		for m := range methods {
			mm := strings.ToUpper(strings.TrimSpace(m))
			if mm == "" {
				mm = "GET"
			}
			out = append(out, model.Endpoint{Method: mm, Path: path})
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
