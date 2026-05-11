package engine

import "strings"

func normalizeTarget(t string) string {
	t = strings.TrimSpace(t)

	if t == "" {
		return ""
	}

	if !strings.HasPrefix(t, "http://") &&
		!strings.HasPrefix(t, "https://") {
		return "https://" + t
	}

	return t
}
