package probe

import "github.com/bspippi1337/restless/internal/core"

func Plan(endpoints []core.Endpoint) []core.Endpoint {
	var out []core.Endpoint

	for _, ep := range endpoints {
		ep.Path = ResolvePath(ep.Path)
		out = append(out, ep)
	}

	return out
}
