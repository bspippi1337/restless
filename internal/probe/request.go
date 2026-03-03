package probe

import "github.com/bspippi1337/restless/internal/core"

type Request struct {
	Method string
	URL    string
}

func Build(base string, ep core.Endpoint) Request {
	path := FillPath(ep.Path)
	return Request{
		Method: ep.Method,
		URL:    base + path,
	}
}
