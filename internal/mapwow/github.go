package mapwow

import "strings"

func enrichGitHub(r *Result) {
	r.Signals = appendUnique(
		r.Signals,
		"github surface",
	)

	r.Signals = appendUnique(
		r.Signals,
		"anonymous api quota",
	)

	if len(r.Nodes) <= 2 {
		r.Signals = appendUnique(
			r.Signals,
			"response degraded by rate limiting",
		)

		r.Topology = appendUnique(
			r.Topology,
			"restricted graph visibility",
		)
	}
}

func inferGitHubTopology(nodes []Node) []string {
	var out []string

	for _, n := range nodes {
		l := strings.ToLower(n.Name)

		switch {
		case strings.Contains(l, "clone"):
			out = appendUnique(
				out,
				"repository replication graph",
			)

		case strings.Contains(l, "fork"):
			out = appendUnique(
				out,
				"fork network topology",
			)

		case strings.Contains(l, "owner"):
			out = appendUnique(
				out,
				"ownership hierarchy",
			)
		}
	}

	return out
}
