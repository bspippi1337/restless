package botanical

import (
	"strings"
)

// resolveParameters expands only locally declared OpenAPI parameter references.
// External URLs, unknown components, cycles and excessive nesting are never fetched.
func resolveParameters(params []Parameter, known map[string]Parameter) ([]Parameter, bool) {
	out := make([]Parameter, 0, len(params))
	unresolved := false
	for _, p := range params {
		visited := make(map[string]bool)
		for depth := 0; p.Ref != "" && depth < 8; depth++ {
			name, ok := localParameterName(p.Ref)
			if !ok || visited[name] {
				unresolved = true
				break
			}
			visited[name] = true
			v, ok := known[name]
			if !ok {
				unresolved = true
				break
			}
			p = v
		}
		if p.Ref != "" {
			unresolved = true
		}
		out = append(out, p)
	}
	return out, unresolved
}

func localParameterName(ref string) (string, bool) {
	const prefix = "#/components/parameters/"
	if !strings.HasPrefix(ref, prefix) {
		return "", false
	}
	s := strings.TrimPrefix(ref, prefix)
	if s == "" || strings.Contains(s, "/") {
		return "", false
	}
	// RFC 6901 JSON Pointer escapes, reject invalid "~" sequences.
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '~' {
			b.WriteByte(s[i])
			continue
		}
		if i+1 >= len(s) {
			return "", false
		}
		i++
		switch s[i] {
		case '0':
			b.WriteByte('~')
		case '1':
			b.WriteByte('/')
		default:
			return "", false
		}
	}
	if b.Len() == 0 {
		return "", false
	}
	return b.String(), true
}
