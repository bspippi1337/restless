package insight

import (
	"sort"
	"strconv"

	"github.com/bspippi1337/restless/internal/core"
)

func LatencyInsights(results []core.EndpointResult) []core.Insight {
	if len(results) == 0 {
		return nil
	}

	var values []int64

	for _, r := range results {
		values = append(values, r.Latency.Milliseconds())
	}

	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })

	p50 := percentile(values, 50)
	p95 := percentile(values, 95)

	var insights []core.Insight

	if p95 > p50*3 {
		insights = append(insights, core.Insight{
			Type: "latency_variance",
			Message: "high latency variance detected",
		})
	}

	insights = append(insights, core.Insight{
		Type: "latency_summary",
		Message: "p50=" + strconv.FormatInt(p50,10) + "ms p95=" + strconv.FormatInt(p95,10) + "ms",
	})

	return insights
}

func percentile(data []int64, p int) int64 {
	if len(data) == 0 {
		return 0
	}

	k := int(float64(len(data)-1) * float64(p) / 100.0)
	return data[k]
}
