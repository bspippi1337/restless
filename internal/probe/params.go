package probe

import (
	"regexp"
	"strings"
)

var pathVar = regexp.MustCompile(`\{([^}]+)\}`)

func FillPath(path string) string {

	return pathVar.ReplaceAllStringFunc(path, func(m string) string {

		name := strings.Trim(m, "{}")

		switch strings.ToLower(name) {

		case "id", "petid", "userid", "orderid":
			return "1"

		case "name", "username":
			return "demo"

		case "status":
			return "available"

		default:
			return "1"
		}
	})
}

func AddQueryDefaults(url string) string {

	if strings.Contains(url, "findByStatus") {
		if !strings.Contains(url, "?") {
			return url + "?status=available"
		}
	}

	if strings.Contains(url, "limit") {
		if !strings.Contains(url, "limit=") {
			if strings.Contains(url, "?") {
				return url + "&limit=10"
			}
			return url + "?limit=10"
		}
	}

	return url
}
