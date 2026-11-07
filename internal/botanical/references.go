package botanical

import (
	"fmt"
	"sort"
	"strings"
)

// ReferenceHit links a local document search result to an immutable archived hash.
// Plain-text references never become executable API candidates.
type ReferenceHit struct {
	Document     Document `json:"document"`
	Score        int      `json:"score"`
	MatchedTerms []string `json:"matched_terms"`
}

func RankReferences(c Catalog, query string, limit int) ([]ReferenceHit, error) {
	if limit < 1 || limit > 50 {
		return nil, fmt.Errorf("limit must be between 1 and 50")
	}
	terms := tokens(query)
	if len(terms) == 0 {
		return nil, fmt.Errorf("query contains no searchable terms")
	}
	var results []ReferenceHit
	for _, doc := range c.Documents {
		name := tokens(doc.Title)
		urlTokens := tokens(doc.URL)
		contentTerms := doc.Terms
		score := 0
		matched := []string{}
		for _, t := range terms {
			weight := 0
			for _, w := range name {
				if w == t {
					weight = 8
					break
				}
			}
			for _, w := range urlTokens {
				if w == t && weight < 5 {
					weight = 5
					break
				}
			}
			for _, w := range contentTerms {
				if w == t && weight < 2 {
					weight = 2
					break
				}
			}
			if weight > 0 {
				score += weight
				matched = append(matched, t)
			}
		}
		if score > 0 {
			if len(matched) == len(terms) {
				score += 10
			}
			if strings.HasPrefix(doc.Kind, "openapi") {
				score += 3
			}
			results = append(results, ReferenceHit{Document: doc, Score: score, MatchedTerms: matched})
		}
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Document.URL < results[j].Document.URL
	})
	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}
