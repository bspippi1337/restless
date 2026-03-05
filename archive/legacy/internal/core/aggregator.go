package core

import "sort"

type Aggregator struct {
	results []EndpointResult
	meta    Meta
}

func NewAggregator() *Aggregator {
	return &Aggregator{
		results: make([]EndpointResult, 0),
	}
}

func (a *Aggregator) Add(r EndpointResult) {
	a.results = append(a.results, r)
}

func (a *Aggregator) SetMeta(m Meta) {
	a.meta = m
}

func (a *Aggregator) Results() []EndpointResult {
	out := make([]EndpointResult, len(a.results))
	copy(out, a.results)

	sort.Slice(out, func(i, j int) bool {
		if out[i].Endpoint.Path == out[j].Endpoint.Path {
			return out[i].Endpoint.Method < out[j].Endpoint.Method
		}
		return out[i].Endpoint.Path < out[j].Endpoint.Path
	})

	return out
}

func (a *Aggregator) Build(specHash, baseURL string) VerificationResult {
	summary := Summary{}

	for _, r := range a.results {
		switch r.Status {
		case StatusOK:
			summary.OK++
		case StatusWarn:
			summary.Warn++
		case StatusFail:
			summary.Fail++
		}
	}

	return VerificationResult{
		SpecHash: specHash,
		BaseURL:  baseURL,
		Results:  a.Results(),
		Summary:  summary,
		Meta:     a.meta,
	}
}
