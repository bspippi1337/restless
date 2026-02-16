package snippets

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/bspippi1337/restless/internal/httpclient"
	"github.com/bspippi1337/restless/internal/profile"
)

type RunOptions struct {
	BaseURLOverride string
	TimeoutSeconds  int
}

func RunSnippet(ctx context.Context, pr profile.Profile, sn Snippet, opt RunOptions) (httpclient.Result, error) {
	base := pr.BaseURLs[0]
	if opt.BaseURLOverride != "" { base = opt.BaseURLOverride }

	headers := map[string]string{}
	for k, v := range pr.Defaults { headers[k] = v }
	for k, v := range sn.Headers { headers[k] = v }

	if strings.ToLower(pr.AuthType) == "bearer" && headers["Authorization"] == "" {
		if tok := os.Getenv(pr.AuthEnv); tok != "" {
			headers["Authorization"] = "Bearer " + tok
		}
	}

	req := httpclient.Request{
		Method:  sn.Method,
		BaseURL: base,
		Path:    sn.Path,
		Headers: headers,
		Query:   map[string]string{},
		Body:    []byte(sn.Body),
	}

	if _, ok := ctx.Deadline(); !ok {
		ts := pr.TimeoutS
		if opt.TimeoutSeconds > 0 { ts = opt.TimeoutSeconds }
		c, cancel := context.WithTimeout(ctx, time.Duration(ts)*time.Second)
		defer cancel()
		return httpclient.Do(c, req)
	}
	return httpclient.Do(ctx, req)
}
