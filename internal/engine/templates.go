package engine

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

var templateURL = regexp.MustCompile(`https?://[^"]+\{[^"]+\}`)

func ExtractTemplatesFromRoot(target string) []string {

	target = normalizeTarget(target)

	client := &http.Client{
		Timeout: 6 * time.Second,
	}

	resp, err := client.Get(target)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var body interface{}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil
	}

	seen := map[string]bool{}
	var templates []string

	walkJSON(body, func(s string) {

		matches := templateURL.FindAllString(s, -1)

		for _, m := range matches {

			u, err := url.Parse(m)
			if err != nil {
				continue
			}

			p := normalizeTemplate(u.Path)

			if p == "" || p == "/" || isDocumentationPath(p) {
				continue
			}

			if !seen[p] {

				seen[p] = true
				templates = append(templates, p)

			}
		}
	})

	return templates
}

func walkJSON(v interface{}, fn func(string)) {

	switch val := v.(type) {

	case map[string]interface{}:

		for _, vv := range val {
			walkJSON(vv, fn)
		}

	case []interface{}:

		for _, vv := range val {
			walkJSON(vv, fn)
		}

	case string:

		fn(val)
	}
}
