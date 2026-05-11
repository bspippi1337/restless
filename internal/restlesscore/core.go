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

	client := &http.Client{Timeout: timeout}

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

	if status >= 200 && status < 500 {
		r.Confirmed = append(r.Confirmed, Endpoint{
			Method:     "GET",
			Path:       "/",
			Status:     status,
			Confidence: "high",
			Source:     "root",
		})
	}

	initial := discover(base, body)
	seen := map[string]bool{}

	for _, p := range initial {
		if seen[p] {
			continue
		}

		seen[p] = true

		status, _, _, _ := fetch(client, base+p)

		r.Confirmed = append(r.Confirmed, Endpoint{
			Method:     "GET",
			Path:       p,
			Status:     status,
			Confidence: "high",
			Source:     "surface",
		})
	}

	r.Confirmed = uniqEndpoints(r.Confirmed)

	sort.Slice(r.Confirmed, func(i, j int) bool {
		return r.Confirmed[i].Path < r.Confirmed[j].Path
	})

	return r, nil
}

func Render(title string, r *ScanResult) string {
	var b strings.Builder

	live := 0
	restricted := 0

	for _, ep := range r.Confirmed {
		if ep.Status >= 200 && ep.Status < 300 {
			live++
		} else {
			restricted++
		}
	}

	fmt.Fprintf(&b, "RESTLESS ENGINE :: %s\n\n", strings.TrimRight(r.Target, "/"))
	fmt.Fprintf(&b, "Type      %s\n", r.APIType)

	if len(r.Fingerprints) > 0 {
		fmt.Fprintf(&b, "Traits    %s\n", strings.Join(r.Fingerprints, " · "))
	}

	fmt.Fprintf(&b, "\n")

	groups := map[string][]Endpoint{}

	for _, ep := range r.Confirmed {
		if ep.Path == "/" {
			continue
		}

		groups[groupName(ep.Path)] = append(groups[groupName(ep.Path)], ep)
	}

	order := []string{
		"Identity",
		"Repositories",
		"Activity",
		"Search",
		"Platform",
		"Misc",
	}

	for _, name := range order {
		items := groups[name]
		if len(items) == 0 {
			continue
		}

		fmt.Fprintf(&b, "%s\n", name)
		fmt.Fprintf(&b, "%s\n", strings.Repeat("-", len(name)))

		for _, ep := range items {
			state := "restricted"
			if ep.Status >= 200 && ep.Status < 300 {
				state = "live"
			}

			fmt.Fprintf(&b, "  %-28s %s\n", ep.Path, state)
		}

		fmt.Fprintf(&b, "\n")
	}

	fmt.Fprintf(&b, "Capability Profile\n")
	fmt.Fprintf(&b, "------------------\n")
	fmt.Fprintf(&b, "  Traversal Surface   %d endpoints\n", len(r.Confirmed))
	fmt.Fprintf(&b, "  Anonymous Access   %d live\n", live)
	fmt.Fprintf(&b, "  Restricted Surface %d gated\n", restricted)
	fmt.Fprintf(&b, "\n")

	fmt.Fprintf(&b, "Suggested Workflows\n")
	fmt.Fprintf(&b, "-------------------\n")
	fmt.Fprintf(&b, "  restless discover %s\n", r.Target)
	fmt.Fprintf(&b, "  restless inspect %s\n", r.Target)
	fmt.Fprintf(&b, "  restless fuzz %s\n", r.Target)
	fmt.Fprintf(&b, "  restless map %s\n", r.Target)

	return b.String()
}

func groupName(path string) string {
	p := strings.ToLower(path)

	switch {
	case strings.Contains(p, "user"):
		return "Identity"
	case strings.Contains(p, "repo"):
		return "Repositories"
	case strings.Contains(p, "event"):
		return "Activity"
	case strings.Contains(p, "search"):
		return "Search"
	case strings.Contains(p, "rate") || strings.Contains(p, "version"):
		return "Platform"
	default:
		return "Misc"
	}
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
	req.Header.Set("User-Agent", "restless-blckswan")
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

	if s := h.Get("Server"); s != "" {
		out = append(out, s)
	}

	if ct := h.Get("Content-Type"); ct != "" {
		ct = strings.ReplaceAll(ct, "application/json; charset=utf-8", "json")
		ct = strings.ReplaceAll(ct, "text/html; charset=utf-8", "html")
		out = append(out, ct)
	}

	if h.Get("X-GitHub-Request-Id") != "" {
		out = append(out, "github-api")
	}

	if h.Get("X-RateLimit-Limit") != "" {
		out = append(out, "rate-limited")
	}

	if json.Valid(body) {
		out = append(out, "json-root")
	}

	return uniqStrings(out)
}

func detectAPIType(fp []string) string {
	for _, f := range fp {
		if strings.Contains(strings.ToLower(f), "github") {
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

			if u.Path != "" && strings.HasPrefix(u.Path, "/") {
				out = append(out, strings.TrimRight(u.Path, "/"))
			}
		}
	}

	out = append(out,
		"/user",
		"/users",
		"/repos",
		"/events",
		"/search",
		"/rate_limit",
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
