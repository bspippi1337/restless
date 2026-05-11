package discovery

import (
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Fingerprint struct {
	Target          string
	APIType         string
	Server          string
	Technologies    []string
	InterestingURLs []string
	GraphQL         bool
	OpenAPI         bool
	Confidence      int
}

func FingerprintTarget(target string) (*Fingerprint, error) {
	if !strings.HasPrefix(target, "http://") &&
		!strings.HasPrefix(target, "https://") {
		target = "https://" + target
	}

	client := &http.Client{
		Timeout: 12 * time.Second,
	}

	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "restless/1337")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	body := string(bodyBytes)

	fp := &Fingerprint{
		Target:       target,
		Server:       resp.Header.Get("Server"),
		Confidence:   15,
		Technologies: []string{},
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))

	if strings.Contains(contentType, "json") {
		fp.APIType = "REST"
		fp.Confidence += 25
	} else {
		fp.APIType = "Web Application"
	}

	if strings.Contains(body, "__schema") ||
		strings.Contains(body, "graphql") {
		fp.GraphQL = true
		fp.Confidence += 20
	}

	openapiHints := []string{
		"openapi",
		"swagger",
		"swagger-ui",
	}

	for _, h := range openapiHints {
		if strings.Contains(strings.ToLower(body), h) {
			fp.OpenAPI = true
			fp.Confidence += 15
			break
		}
	}

	techPatterns := map[string]string{
		"cloudflare": "Cloudflare",
		"react":      "React",
		"vue":        "Vue",
		"angular":    "Angular",
		"next.js":    "Next.js",
		"nuxt":       "Nuxt",
		"jquery":     "jQuery",
		"wordpress":  "WordPress",
		"drupal":     "Drupal",
		"graphql":    "GraphQL",
	}

	lowerBody := strings.ToLower(body)

	for k, v := range techPatterns {
		if strings.Contains(lowerBody, k) {
			fp.Technologies = append(fp.Technologies, v)
		}
	}

	urlRegex := regexp.MustCompile(`(?i)(/api/[a-zA-Z0-9_\-/]+)`)
	matches := urlRegex.FindAllString(body, 50)

	seen := map[string]bool{}

	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			fp.InterestingURLs = append(fp.InterestingURLs, m)
		}
	}

	if fp.GraphQL {
		fp.APIType = "REST + GraphQL"
	}

	return fp, nil
}
