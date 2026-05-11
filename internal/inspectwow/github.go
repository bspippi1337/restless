package inspectwow

import (
	"encoding/json"
	"net/http"
)

func enrichGitHub(
	req *http.Request,
	resp *http.Response,
	r *Result,
	body []byte,
) {
	if resp.StatusCode == 403 {
		if rem := resp.Header.Get("X-RateLimit-Remaining"); rem == "0" {
			r.Problems = appendUnique(
				r.Problems,
				"github rate limit exceeded",
			)

			r.Signals = appendUnique(
				r.Signals,
				"anonymous github quota exhausted",
			)

			r.Examples = append(
				r.Examples,
				`export GITHUB_TOKEN=xxxxx`,
			)

			r.Examples = append(
				r.Examples,
				`restless inspect https://api.github.com/repos/torvalds/linux -H "Authorization: Bearer $GITHUB_TOKEN"`,
			)
		}
	}

	var obj map[string]interface{}

	if json.Unmarshal(body, &obj) == nil {
		if _, ok := obj["stargazers_count"]; ok {
			r.Relations = appendUnique(
				r.Relations,
				"repository",
			)

			r.Signals = appendUnique(
				r.Signals,
				"public repository surface",
			)
		}
	}
}
