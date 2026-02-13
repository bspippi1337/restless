package model

import "time"

type EvidenceSource string

const (
	SourceOpenAPI EvidenceSource = "openapi"
	SourceSitemap EvidenceSource = "sitemap"
	SourceRobots  EvidenceSource = "robots"
	SourceHTML    EvidenceSource = "html"
	SourceProbe   EvidenceSource = "probe"
	SourceFuzzer  EvidenceSource = "fuzzer"
	SourceOther   EvidenceSource = "other"
)

type Evidence struct {
	Source EvidenceSource `json:"source,omitempty"`
	URL    string         `json:"url,omitempty"`
	Note   string         `json:"note,omitempty"`
	When   time.Time      `json:"when,omitempty"`
	Score  float64        `json:"score,omitempty"`
}

type Endpoint struct {
	Method    string     `json:"method,omitempty"`
	Path      string     `json:"path,omitempty"`
	FullURL   string     `json:"fullUrl,omitempty"`
	Evidences []Evidence `json:"evidences,omitempty"`
}

type Finding struct {
	BaseURL    string     `json:"baseUrl,omitempty"`
	Hosts      []string   `json:"hosts,omitempty"`
	DocURLs    []string   `json:"docUrls,omitempty"`
	Endpoints  []Endpoint `json:"endpoints,omitempty"`
	Notes      []string   `json:"notes,omitempty"`
	Confidence float64    `json:"confidence,omitempty"`
}
