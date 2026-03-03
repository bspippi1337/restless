package insight

import (
	"github.com/bspippi1337/restless/internal/core"
)

func DriftInsights(prev, curr []core.EndpointResult) []core.Insight {
	if len(prev) == 0 {
		return nil
	}

	m := map[string]core.Status{}

	for _, r := range prev {
		key := r.Endpoint.Method + " " + r.Endpoint.Path
		m[key] = r.Status
	}

	var insights []core.Insight

	for _, r := range curr {
		key := r.Endpoint.Method + " " + r.Endpoint.Path

		if old, ok := m[key]; ok && old != r.Status {
			insights = append(insights, core.Insight{
				Type: "drift",
				Message: key + " changed from " + string(old) + " to " + string(r.Status),
			})
		}
	}

	return insights
}
