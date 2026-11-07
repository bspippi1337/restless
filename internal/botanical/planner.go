package botanical

import (
	"fmt"
	"sort"
	"strings"
)

type Hit struct {
	Operation     Operation `json:"operation"`
	Score         int       `json:"score"`
	MatchedFields []string  `json:"matched_fields"`
}
type ProbeStep struct {
	Rank                 int                   `json:"rank"`
	Method               string                `json:"method"`
	DocumentedPath       string                `json:"documented_path"`
	OperationID          string                `json:"operation_id,omitempty"`
	Purpose              string                `json:"purpose"`
	Parameters           []Parameter           `json:"parameters,omitempty"`
	Security             []map[string][]string `json:"security,omitempty"`
	SourceURL            string                `json:"source_url"`
	SourceSHA256         string                `json:"source_sha256"`
	Confidence           string                `json:"confidence"`
	RequiresManualReview bool                  `json:"requires_manual_review"`
	ReviewReasons        []string              `json:"review_reasons,omitempty"`
	Score                int                   `json:"relevance_score"`
}
type ProbePlan struct {
	Query           string      `json:"query"`
	Mode            string      `json:"mode"`
	NetworkRequests int         `json:"network_requests"`
	Candidates      []ProbeStep `json:"candidates"`
	Notes           []string    `json:"notes"`
}

func RankOperations(c Catalog, query string, maxResults int) ([]Hit, error) {
	if maxResults < 1 || maxResults > 50 {
		return nil, fmt.Errorf("limit must be between 1 and 50")
	}
	wanted := tokens(query)
	if len(wanted) == 0 {
		return nil, fmt.Errorf("query contains no searchable terms")
	}
	mutatingIntent := false
	for _, term := range wanted {
		switch term {
		case "create", "add", "put", "post", "update", "patch", "delete", "remove", "upload", "write":
			mutatingIntent = true
		}
	}
	var hits []Hit
	for _, op := range c.Operations {
		groups := []struct {
			label, text string
			weight      int
		}{
			{"operation_id", op.OperationID, 8},
			{"path", op.Path, 6},
			{"summary", op.Summary, 6},
			{"description", op.Description, 2},
			{"tags", strings.Join(op.Tags, " "), 4},
			{"method", op.Method, 3},
		}
		for _, param := range op.Parameters {
			groups = append(groups, struct {
				label, text string
				weight      int
			}{"parameter", param.Name + " " + param.Description, 2})
		}
		score := 0
		matched := []string{}
		covered := map[string]bool{}
		for _, group := range groups {
			groupTokens := tokens(group.text)
			matches := 0
			for _, w := range wanted {
				for _, candidate := range groupTokens {
					if candidate == w || (len(w) > 3 && strings.HasPrefix(candidate, w)) {
						covered[w] = true
						matches++
						break
					}
				}
			}
			if matches > 0 {
				score += matches * group.weight
				matched = append(matched, group.label)
			}
		}
		if score == 0 {
			continue
		}
		score += len(covered) * 10
		if len(covered) == len(wanted) {
			score += 15
		}
		if !mutatingIntent && (op.Method == "GET" || op.Method == "HEAD") {
			score += 4
		}
		hits = append(hits, Hit{Operation: op, Score: score, MatchedFields: matched})
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		if hits[i].Operation.SourceURL != hits[j].Operation.SourceURL {
			return hits[i].Operation.SourceURL < hits[j].Operation.SourceURL
		}
		if hits[i].Operation.Path != hits[j].Operation.Path {
			return hits[i].Operation.Path < hits[j].Operation.Path
		}
		return hits[i].Operation.Method < hits[j].Operation.Method
	})
	if len(hits) > maxResults {
		hits = hits[:maxResults]
	}
	return hits, nil
}

func BuildProbePlan(c Catalog, query string, maxResults int) (ProbePlan, error) {
	p := ProbePlan{
		Query: query, Mode: "DOCUMENTATION_ONLY_NO_EXECUTION", NetworkRequests: 0,
		Notes: []string{
			"Only operations explicitly present in locally archived OpenAPI 3 specifications are considered.",
			"Paths are templates, not verified live endpoints. No URLs are resolved, requests generated, or probes executed.",
			"Authentication requirements are copied from the specification, not validated against a server.",
			"Check owner permission, approved target scope and operational impact before any live API testing.",
		},
	}
	hits, err := RankOperations(c, query, maxResults)
	if err != nil {
		return p, err
	}
	for i, h := range hits {
		o := h.Operation
		s := ProbeStep{
			Rank: i + 1, Method: o.Method, DocumentedPath: o.Path, OperationID: o.OperationID,
			Purpose: o.Summary, Parameters: o.Parameters, Security: o.Security,
			SourceURL: o.SourceURL, SourceSHA256: o.SourceSHA256, Score: h.Score, Confidence: "documented",
		}
		if s.Purpose == "" {
			s.Purpose = o.Description
		}
		if o.Method != "GET" && o.Method != "HEAD" {
			s.RequiresManualReview = true
			s.ReviewReasons = append(s.ReviewReasons, "Non-read HTTP method: do not execute automatically")
		}
		if o.AuthSpecified {
			s.RequiresManualReview = true
			s.ReviewReasons = append(s.ReviewReasons, "Authentication or authorization required")
		}
		if strings.Contains(o.Path, "{") {
			s.RequiresManualReview = true
			s.ReviewReasons = append(s.ReviewReasons, "Path parameters require explicit authorized values")
		}
		for _, param := range o.Parameters {
			if param.Required {
				s.RequiresManualReview = true
				s.ReviewReasons = append(s.ReviewReasons, "Required "+param.In+" parameter: "+param.Name)
			}
		}
		if o.UnresolvedRefs {
			s.RequiresManualReview = true
			s.ReviewReasons = append(s.ReviewReasons, "Unresolved OpenAPI $ref: manual specification review")
			s.Confidence = "partial"
		}
		p.Candidates = append(p.Candidates, s)
	}
	if len(p.Candidates) == 0 {
		p.Notes = append(p.Notes, "No documented operation matched the query; no endpoint was invented.")
	}
	return p, nil
}
