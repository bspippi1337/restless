package engine

import "strings"

func classifyEndpoint(path string) string {
	switch {
	case strings.Contains(path, "graphql"):
		return "graphql"
	case strings.Contains(path, "search"):
		return "search"
	case strings.Contains(path, "status"), strings.Contains(path, "health"), strings.Contains(path, "rate_limit"):
		return "health"
	case strings.Contains(path, "{"):
		return "resource"
	default:
		return "core"
	}
}

func normalizeTemplate(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}

	p = strings.ReplaceAll(p, "{/", "/{")
	p = strings.ReplaceAll(p, "//", "/")
	p = strings.TrimSuffix(p, "{")
	p = strings.TrimSpace(p)

	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}

	for strings.Contains(p, "//") {
		p = strings.ReplaceAll(p, "//", "/")
	}

	return p
}

func isDocumentationPath(p string) bool {
	docTokens := []string{
		"/docs",
		"/documentation",
		"/guides",
		"/guide",
		"/overview",
		"/rest/overview",
	}

	for _, t := range docTokens {
		if strings.Contains(p, t) {
			return true
		}
	}
	return false
}
