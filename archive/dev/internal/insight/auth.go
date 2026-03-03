package insight

import "github.com/bspippi1337/restless/internal/core"

func AuthInsights(results []core.EndpointResult) []core.Insight {
	if len(results) == 0 {
		return nil
	}

	var total int
	var unauthorized int

	for _, r := range results {
		total++
		if r.HTTPStatus == 401 {
			unauthorized++
		}
	}

	if unauthorized == 0 {
		return nil
	}

	ratio := float64(unauthorized) / float64(total)

	if ratio > 0.5 {
		return []core.Insight{
			{
				Type:    "auth_missing",
				Message: "most endpoints returned 401 unauthorized; authentication may be required",
			},
		}
	}

	return nil
}
