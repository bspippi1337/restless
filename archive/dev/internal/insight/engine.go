package insight

import "github.com/bspippi1337/restless/internal/core"

func Analyze(results []core.EndpointResult) []core.Insight {
	var out []core.Insight

	out = append(out, LatencyInsights(results)...)
	out = append(out, AuthInsights(results)...)

	return out
}
