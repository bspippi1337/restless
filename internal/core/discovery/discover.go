package discovery

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Options struct {
	BudgetSeconds int
	BudgetPages   int
	Verify        bool
	Fuzz          bool
	Debug         bool
}

type Finding struct {
	Domain     string     `json:"domain"`
	BaseURLs   []string   `json:"baseUrls"`
	DocURLs    []string   `json:"docUrls"`
	Endpoints  []Endpoint `json:"endpoints"`
	Confidence float64    `json:"confidence"`
}

type Endpoint struct {
	Method   string     `json:"method"`
	Path     string     `json:"path"`
	Score    float64    `json:"score"`
	Evidence []Evidence `json:"evidence"`
}

type Evidence struct {
	Source string  `json:"source"`
	URL    string  `json:"url"`
	When   string  `json:"when"`
	Score  float64 `json:"score"`
}

func DiscoverDomain(domain string, opt Options) (Finding, error) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return Finding{}, errors.New("empty domain")
	}
	if opt.BudgetSeconds <= 0 {
		opt.BudgetSeconds = 15
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(opt.BudgetSeconds)*time.Second)
	defer cancel()

	now := time.Now().Format(time.RFC3339)

	find := Finding{
		Domain:     domain,
		BaseURLs:   []string{fmt.Sprintf("https://api.%s", domain)},
		DocURLs:    []string{fmt.Sprintf("https://%s/openapi.json", domain)},
		Confidence: 0.50,
		Endpoints: []Endpoint{
			{
				Method: "GET",
				Path:   "/v1/status",
				Score:  0.50,
				Evidence: []Evidence{
					{Source: "heuristic", URL: fmt.Sprintf("https://%s/", domain), When: now, Score: 0.50},
				},
			},
		},
	}

	// Optional verify: cheap HEAD/GET check for base URL root
	if opt.Verify {
		u := fmt.Sprintf("https://%s/", domain)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		resp, err := http.DefaultClient.Do(req)
		if err == nil && resp != nil {
			_ = resp.Body.Close()
			// bump confidence if any response came back
			find.Confidence = 0.65
			find.Endpoints[0].Score = 0.65
			find.Endpoints[0].Evidence = append(find.Endpoints[0].Evidence, Evidence{
				Source: "verify",
				URL:    u,
				When:   now,
				Score:  0.65,
			})
		}
	}

	_ = ctx // silence linters if future changes remove verify usage

	return find, nil
}
