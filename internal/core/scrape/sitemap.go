package scrape

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type urlset struct {
	URLs []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
}

func SitemapDocs(ctx context.Context, base string, maxURLs int) ([]string, []string) {
	u := strings.TrimRight(base, "/") + "/sitemap.xml"
	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	req.Header.Set("User-Agent", "restless-sitemap/0.2")
	res, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, nil
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, nil
	}
	b, _ := io.ReadAll(io.LimitReader(res.Body, 3<<20))

	var xs urlset
	if err := xml.Unmarshal(b, &xs); err != nil {
		return []string{u}, nil
	}

	paths := []string{}
	for _, it := range xs.URLs {
		if len(paths) >= maxURLs {
			break
		}
		loc := strings.TrimSpace(it.Loc)
		if loc == "" {
			continue
		}
		pu, err := url.Parse(loc)
		if err != nil || pu.Path == "" {
			continue
		}
		p := pu.Path
		if strings.Contains(p, "/api") || strings.HasPrefix(p, "/v1/") || strings.Contains(p, "/swagger") || strings.Contains(p, "/openapi") {
			paths = append(paths, p)
		}
	}
	return []string{u}, paths
}
