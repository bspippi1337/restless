package restlesscore

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Endpoint struct {
	Method     string
	Path       string
	Status     int
	Confidence string
	Source     string
}

type Edge struct {
	From string
	To   string
}

type ScanResult struct {
	Target       string
	BaseURL      string
	APIType      string
	Fingerprints []string
	Confirmed    []Endpoint
	Topology     []Edge
}

type CrawlNode struct {
	Path  string
	Depth int
}

func Scan(target string, timeout time.Duration) (*ScanResult, error) {

	base := normalize(target)

	client := &http.Client{
		Timeout: timeout,
	}

	r := &ScanResult{
		Target:  target,
		BaseURL: base,
	}

	status, body, headers, err := fetch(client, base+"/")
	if err != nil {
		return nil, err
	}

	r.Fingerprints = fingerprints(headers, body)
	r.APIType = detectAPIType(r.Fingerprints)

	if status >= 200 && status < 400 {
		r.Confirmed = append(r.Confirmed, Endpoint{
			Method:     "GET",
			Path:       "/",
			Status:     status,
			Confidence: "high",
			Source:     "root",
		})
	}

	initial := discover(base, body)

	queue := []CrawlNode{}

	for _, p := range initial {
		queue = append(queue, CrawlNode{
			Path:  p,
			Depth: 1,
		})
	}

	seen := map[string]bool{}

	for len(queue) > 0 {

		node := queue[0]
		queue = queue[1:]

		if seen[node.Path] {
			continue
		}

		seen[node.Path] = true

		status, body, _, err := fetch(client, base+node.Path)

		ep := Endpoint{
			Method: "GET",
			Path:   node.Path,
			Status: status,
			Source: "crawl",
		}

		if err == nil {

			switch {

			case status >= 200 && status < 300:
				ep.Confidence = "high"
				r.Confirmed = append(r.Confirmed, ep)

			case status == 401 || status == 403 || status == 405:
				ep.Confidence = "medium"
				r.Confirmed = append(r.Confirmed, ep)
			}

			r.Topology = append(r.Topology, Edge{
				From: "/",
				To:   node.Path,
			})

			if node.Depth < 2 {

				next := discover(base, body)

				for _, np := range next {

					if !seen[np] {

						r.Topology = append(r.Topology, Edge{
							From: node.Path,
							To:   np,
						})

						queue = append(queue, CrawlNode{
							Path:  np,
							Depth: node.Depth + 1,
						})
					}
				}
			}
		}
	}

	r.Confirmed = uniqEndpoints(r.Confirmed)
	r.Topology = uniqEdges(r.Topology)

	sort.Slice(r.Confirmed, func(i, j int) bool {
		return r.Confirmed[i].Path < r.Confirmed[j].Path
	})

	return r, nil
}

func Render(title string, r *ScanResult) string {

	var b strings.Builder

	type bucket struct {
		Name  string
		Icon  string
		Paths []Endpoint
	}

	buckets := map[string]*bucket{
		"identity": {
			Name: "Identity",
			Icon: "👤",
		},
		"repositories": {
			Name: "Repositories",
			Icon: "📦",
		},
		"activity": {
			Name: "Activity",
			Icon: "📡",
		},
		"search": {
			Name: "Search",
			Icon: "🔍",
		},
		"platform": {
			Name: "Platform",
			Icon: "⏱",
		},
		"misc": {
			Name: "Misc",
			Icon: "•",
		},
	}

	for _, ep := range r.Confirmed {

		if ep.Path == "/" {
			continue
		}

		switch {

		case strings.Contains(ep.Path, "user"):
			buckets["identity"].Paths =
				append(buckets["identity"].Paths, ep)

		case strings.Contains(ep.Path, "repo"):
			buckets["repositories"].Paths =
				append(buckets["repositories"].Paths, ep)

		case strings.Contains(ep.Path, "event"):
			buckets["activity"].Paths =
				append(buckets["activity"].Paths, ep)

		case strings.Contains(ep.Path, "search"):
			buckets["search"].Paths =
				append(buckets["search"].Paths, ep)

		case strings.Contains(ep.Path, "rate"):
			buckets["platform"].Paths =
				append(buckets["platform"].Paths, ep)

		default:
			buckets["misc"].Paths =
				append(buckets["misc"].Paths, ep)
		}
	}

	live := 0
	restricted := 0

	for _, ep := range r.Confirmed {

		if ep.Status >= 200 && ep.Status < 300 {
			live++
		} else {
			restricted++
		}
	}

	fmt.Fprintf(
		&b,
		title,
		r.Target,
	)

	fmt.Fprintf(&b, "Type  %s\n", r.APIType)

	if len(r.Fingerprints) > 0 {

		fmt.Fprintf(&b, "Traits  ")

		for i, fp := range r.Fingerprints {

			if i > 0 {
				fmt.Fprintf(&b, " · ")
			}

			fp = strings.ReplaceAll(fp, "application/json; charset=utf-8", "json")
			fp = strings.ReplaceAll(fp, "text/html; charset=utf-8", "html")

			fmt.Fprintf(&b, "%s", fp)
		}

		fmt.Fprintf(&b, "\n")
	}

	fmt.Fprintf(&b, "\n")

	order := []string{
		"identity",
		"repositories",
		"activity",
		"search",
		"platform",
		"misc",
	}

	for _, key := range order {

		bk := buckets[key]

		if len(bk.Paths) == 0 {
			continue
		}

		fmt.Fprintf(
			&b,
			"%s %s\n",
			bk.Icon,
			bk.Name,
		)

		for _, ep := range bk.Paths {

			state := "restricted"

			if ep.Status >= 200 && ep.Status < 300 {
				state = "live"
			}

			fmt.Fprintf(
				&b,
				"  %-24s %s\n",
				ep.Path,
				state,
			)
		}

		fmt.Fprintf(&b, "\n")
	}

	fmt.Fprintf(&b, "Graph\n")
	fmt.Fprintf(&b, "-----\n")

	fmt.Fprintf(&b, "  /\n")

	for _, key := range order {

		bk := buckets[key]

		if len(bk.Paths) == 0 {
			continue
		}

		fmt.Fprintf(
			&b,
			"  - %s\n",
			strings.ToLower(bk.Name),
		)

		for _, ep := range bk.Paths {

			fmt.Fprintf(
				&b,
				"  - %s\n",
				ep.Path,
			)
		}
	}

	fmt.Fprintf(&b, "\n")

	fmt.Fprintf(&b, "Surface\n")
	fmt.Fprintf(&b, "-------\n")

	fmt.Fprintf(
		&b,
		"  %d discovered · %d live · %d restricted\n",
		len(r.Confirmed),
		live,
		restricted,
	)

	return b.String()
}

func normalize(raw string) string {

	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "https://")
	raw = strings.TrimPrefix(raw, "http://")
	raw = strings.TrimRight(raw, "/")

	return "https://" + raw
}

func fetch(client *http.Client, u string) (int, []byte, http.Header, error) {

	req, _ := http.NewRequest("GET", u, nil)

	req.Header.Set("User-Agent", "restless-crawler")
	req.Header.Set("Accept", "application/json,text/html,*/*")

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))

	return resp.StatusCode, body, resp.Header, nil
}

func fingerprints(h http.Header, body []byte) []string {

	var out []string

	if ct := h.Get("Content-Type"); ct != "" {
		out = append(out, ct)
	}

	if h.Get("X-GitHub-Request-Id") != "" {
		out = append(out, "github-api")
	}

	if h.Get("X-RateLimit-Limit") != "" {
		out = append(out, "rate-limited")
	}

	if s := h.Get("Server"); s != "" {
		out = append(out, s)
	}

	if json.Valid(body) {
		out = append(out, "json-root")
	}

	return uniqStrings(out)
}

func detectAPIType(fp []string) string {

	for _, f := range fp {

		if strings.Contains(f, "github-api") {
			return "REST catalog"
		}
	}

	return "REST/JSON"
}

func discover(base string, body []byte) []string {

	var out []string

	var data map[string]any

	if json.Unmarshal(body, &data) == nil {

		for _, v := range data {

			s, ok := v.(string)
			if !ok {
				continue
			}

			if strings.Contains(s, "{") {
				s = strings.Split(s, "{")[0]
			}

			u, err := url.Parse(s)
			if err != nil {
				continue
			}

			if u.Host != "" {

				baseURL, _ := url.Parse(base)

				if baseURL.Host != u.Host {
					continue
				}
			}

			if u.Path != "" && strings.HasPrefix(u.Path, "/") {
				out = append(out, strings.TrimRight(u.Path, "/"))
			}
		}
	}

	out = append(out,
		"/rate_limit",
		"/users",
		"/repos",
		"/search",
		"/events",
		"/user",
	)

	return uniqStrings(out)
}

func uniqStrings(in []string) []string {

	m := map[string]bool{}
	var out []string

	for _, s := range in {

		if s == "" || m[s] {
			continue
		}

		m[s] = true
		out = append(out, s)
	}

	sort.Strings(out)

	return out
}

func uniqEndpoints(in []Endpoint) []Endpoint {

	m := map[string]bool{}
	var out []Endpoint

	for _, e := range in {

		k := e.Method + " " + e.Path

		if m[k] {
			continue
		}

		m[k] = true
		out = append(out, e)
	}

	return out
}

func uniqEdges(in []Edge) []Edge {

	m := map[string]bool{}
	var out []Edge

	for _, e := range in {

		k := e.From + "->" + e.To

		if m[k] {
			continue
		}

		m[k] = true
		out = append(out, e)
	}

	return out
}
