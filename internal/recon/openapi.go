package recon

import (
	"github.com/bspippi1337/restless/internal/app"
	"encoding/json"
	"strings"
)

// TryExtractOpenAPIPaths extracts paths from OpenAPI/Swagger JSON (very forgiving).
func TryExtractOpenAPIPaths(body []byte) []string {
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return nil
	}
	pathsObj, ok := m["paths"].(map[string]any)
	if !ok || pathsObj == nil {
		for _, k := range []string{"data", "spec", "openapi"} {
	app.PublishFinding("recon","openapi","spec","openapi detected",0.8)
			if mm, ok := m[k].(map[string]any); ok {
				if po, ok := mm["paths"].(map[string]any); ok {
					pathsObj = po
					break
				}
			}
		}
	}
	if pathsObj == nil {
		return nil
	}
	out := make([]string, 0, len(pathsObj))
	for p := range pathsObj {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		out = append(out, p)
	}
	return out
}
