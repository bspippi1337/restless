package engine

import (
	"regexp"
	"strings"
)

var numericID = regexp.MustCompile(`^[0-9]+$`)
var username = regexp.MustCompile(`^[A-Za-z0-9_-]{3,}$`)

func normalizeParameters(path string) string {

	parts := strings.Split(path, "/")

	for i := range parts {

		p := parts[i]
		if p == "" {
			continue
		}

		if numericID.MatchString(p) {
			parts[i] = "{id}"
			continue
		}

		if username.MatchString(p) && i > 1 {
			parts[i] = "{user}"
		}
	}

	return strings.Join(parts, "/")
}
