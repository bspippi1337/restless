package runtime

import (
	"regexp"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// MatchPathTemplate attempts to map a concrete request path (e.g. /users/42)
// to an OpenAPI template path (e.g. /users/{id}).
// Returns the matching template and true if found.
func MatchPathTemplate(doc *openapi3.T, concretePath string) (string, bool) {
	if doc == nil || doc.Paths == nil {
		return "", false
	}

	// Prefer longer (more specific) templates first.
	templates := make([]string, 0, len(doc.Paths.Map()))
	for t := range doc.Paths.Map() {
		templates = append(templates, t)
	}
	sort.Slice(templates, func(i, j int) bool { return len(templates[i]) > len(templates[j]) })

	for _, t := range templates {
		if t == concretePath {
			return t, true
		}
		re := regexp.MustCompile(templateToRegex(t))
		if re.MatchString(concretePath) {
			return t, true
		}
	}
	return "", false
}

func templateToRegex(t string) string {
	// Convert /foo/{bar}/baz -> ^/foo/[^/]+/baz$
	r := regexp.MustCompile(`\{[^/]+\}`)
	s := r.ReplaceAllString(t, `[^/]+`)
	s = strings.ReplaceAll(s, "//", "/")
	return "^" + s + "$"
}
