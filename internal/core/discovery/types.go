package discovery

import "time"

type SourceType string

const (
	SourceOpenAPI   SourceType = "openapi"
	SourceSitemap   SourceType = "sitemap"
	SourceHTML      SourceType = "html"
	SourceFuzzer    SourceType = "fuzzer"
	SourceProbe     SourceType = "probe"
)

type Evidence struct {
	Source SourceType `json:"source"`
	URL    string     `json:"url,omitempty"`
	Note   string     `json:"note,omitempty"`
	When   time.Time  `json:"when"`
	Score  int        `json:"score"`
}

type Endpoint struct {
	Method    string     `json:"method"`
	Path      string     `json:"path"`
	FullURL   string     `json:"fullUrl,omitempty"`
	Evidences []Evidence `json:"evidences,omitempty"`
}

type Finding struct {
	BaseURL   string     `json:"baseUrl"`
	DocURLs   []string   `json:"docUrls,omitempty"`
	Hosts     []string   `json:"hosts,omitempty"`
	Endpoints []Endpoint `json:"endpoints,omitempty"`
	Notes     []string   `json:"notes,omitempty"`
}
