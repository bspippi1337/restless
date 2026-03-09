package engine

import "net/http"

func DetectAPIType(target string) string {

	resp, err := http.Get(target)
	if err != nil {
		return "unknown"
	}

	defer resp.Body.Close()

	if resp.Header.Get("X-GitHub-Media-Type") != "" {
		return "REST + GraphQL"
	}

	if resp.Header.Get("Content-Type") == "application/graphql" {
		return "GraphQL"
	}

	return "REST"
}
