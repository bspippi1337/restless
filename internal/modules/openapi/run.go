package openapi

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/bspippi1337/restless/internal/core/types"
)

type RunArgs struct {
	ID     string
	Method string
	Path   string

	BaseOverride string
	PathParams   map[string]string
	QueryParams  map[string]string
	Headers      map[string]string

	Body       []byte
	ShowCurl   bool
	SaveAsName string
}

func BuildRequest(idx SpecIndex, spec Spec, ra RunArgs) (types.Request, string, error) {
	method := strings.ToUpper(strings.TrimSpace(ra.Method))
	if method == "" {
		return types.Request{}, "", errors.New("missing method")
	}
	path := ra.Path
	if path == "" {
		return types.Request{}, "", errors.New("missing path")
	}

	base := strings.TrimRight(spec.BaseURL(), "/")
	if ra.BaseOverride != "" {
		base = strings.TrimRight(ra.BaseOverride, "/")
	}
	if base == "" {
		return types.Request{}, "", errors.New("missing base url (spec has no servers[0].url and no --base provided)")
	}

	// Replace {param} in path
	for k, v := range ra.PathParams {
		path = strings.ReplaceAll(path, "{"+k+"}", url.PathEscape(v))
	}

	full := base + path

	// Query params
	if len(ra.QueryParams) > 0 {
		u, err := url.Parse(full)
		if err != nil {
			return types.Request{}, "", err
		}
		q := u.Query()
		for k, v := range ra.QueryParams {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
		full = u.String()
	}

	h := mapToHeaderKV(ra.Headers)

	req := types.Request{
		Method:  method,
		URL:     full,
		Headers: h,
		Body:    ra.Body,
	}

	curl := ""
	if ra.ShowCurl {
		curl = CurlSnippet(req)
	}

	return req, curl, nil
}

func mapToHeaderKV(m map[string]string) map[string][]string {
	out := map[string][]string{}
	for k, v := range m {
		out[k] = []string{v}
	}
	return out
}

func CurlSnippet(req types.Request) string {
	var b strings.Builder
	b.WriteString("curl -i")
	b.WriteString(" -X ")
	b.WriteString(shellEscape(req.Method))
	for k, vv := range req.Headers {
		for _, v := range vv {
			b.WriteString(" -H ")
			b.WriteString(shellEscape(fmt.Sprintf("%s: %s", k, v)))
		}
	}
	if len(req.Body) > 0 {
		b.WriteString(" --data ")
		b.WriteString(shellEscape(string(req.Body)))
	}
	b.WriteString(" ")
	b.WriteString(shellEscape(req.URL))
	return b.String()
}

// minimal POSIX-ish single-quote escape
func shellEscape(s string) string {
	if s == "" {
		return "''"
	}
	// Wrap in single quotes; escape embedded single quotes: ' -> '"'"'
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}
