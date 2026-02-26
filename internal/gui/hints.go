package gui

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func BuildHints(currentURL string, raw string) (smart []string, links []string) {
	links = extractLinks(raw)

	var v any
	if json.Unmarshal([]byte(raw), &v) == nil {
		smart = smartFromJSON(currentURL, v)
	}
	return smart, links
}

func extractLinks(raw string) []string {
	re := regexp.MustCompile(`https?://[^\s"'<>]+`)
	return re.FindAllString(raw, 50)
}

func smartFromJSON(base string, v any) []string {
	u, err := url.Parse(base)
	if err != nil {
		return nil
	}
	var out []string

	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			kl := strings.ToLower(k)
			if kl == "id" || strings.HasSuffix(kl, "_id") {
				switch vv := val.(type) {
				case float64:
					out = append(out, joinPath(u, strconv.Itoa(int(vv))))
				case string:
					out = append(out, joinPath(u, vv))
				}
			}
		}
	case []any:
		if len(t) > 0 {
			return smartFromJSON(base, t[0])
		}
	}
	return out
}

func joinPath(u *url.URL, seg string) string {
	uu := *u
	uu.RawQuery = ""
	uu.Path = strings.TrimRight(u.Path, "/") + "/" + url.PathEscape(seg)
	return uu.String()
}
