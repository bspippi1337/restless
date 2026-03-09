package engine

import (
	"bytes"
	"net/http"
	"strings"
	"time"
)

func normalizeTarget(t string) string {

	if strings.HasPrefix(t, "http://") || strings.HasPrefix(t, "https://") {
		return t
	}

	return "https://" + t
}

func DetectAPIType(target string) string {

	target = normalizeTarget(target)

	client := &http.Client{
		Timeout: 6 * time.Second,
	}

	apiType := "unknown"

	// ----- REST detection -----

	resp, err := client.Get(target)
	if err == nil {

		ct := resp.Header.Get("Content-Type")

		if strings.Contains(ct, "json") {
			apiType = "REST"
		}

		if resp.Header.Get("X-GitHub-Media-Type") != "" {
			apiType = "REST"
		}

		resp.Body.Close()
	}

	// ----- GraphQL detection -----

	gql := target + "/graphql"

	query := []byte(`{"query":"{__typename}"}`)

	req, _ := http.NewRequest("POST", gql, bytes.NewBuffer(query))
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)

	if err == nil {

		if resp.StatusCode < 500 {

			if apiType == "REST" {
				return "REST + GraphQL"
			}

			return "GraphQL"
		}

		resp.Body.Close()
	}

	return apiType
}
