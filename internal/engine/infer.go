package engine

func inferResources(endpoints *[]Endpoint, seen map[string]bool) {

	candidates := []struct {
		base    string
		pattern string
	}{
		{"/users", "/users/{user}"},
		{"/repos", "/repos/{owner}/{repo}"},
		{"/repos", "/repos/{owner}/{repo}/issues"},
		{"/orgs", "/orgs/{org}"},
		{"/orgs", "/orgs/{org}/repos"},
	}

	for _, c := range candidates {

		if seen[c.base] && !seen[c.pattern] {

			seen[c.pattern] = true

			*endpoints = append(*endpoints, Endpoint{
				Path:       c.pattern,
				Confidence: "high",
			})
		}
	}
}
