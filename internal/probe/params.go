package probe

import (
	"regexp"
)

var paramRE = regexp.MustCompile(`\{([^}]+)\}`)

func ResolvePath(path string) string {
	return paramRE.ReplaceAllString(path, "1")
}
