package util

import "strings"

func JoinURL(base, path string) string {

	base = strings.TrimRight(base, "/")

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return base + path
}
