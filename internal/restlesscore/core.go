package restlesscore

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
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
	From  string
	To    string
	Label string
}

type ScanResult struct {
	Target       string
	BaseURL      string
	APIType      string
	Fingerprints []string
	Candidates   []string
	Confirmed    []Endpoint
	Rejected     []Endpoint
	Topology     []Edge
	Notes        []string
}

func Scan(target string, timeout time.Duration) (*ScanResult, error) {

	base, err := normalize(target)
	if err != nil {
		return nil, err
	}

	c := &http.Client{
		Timeout: timeout,
	}

	r := &ScanResult{
		Target:  target,
		BaseURL: base,
		APIType: "unknown",
	}

	status, body, headers, err := fetch(c, base+"/")
	if err != nil {
		return nil, err
	}

	r.Fingerprints = fingerprints(headers, body)
	r.APIType = detectAPIType(r.Fingerprints, body)

	if status >= 200 && status < 400 {
		r.Confirmed = append(r.Confirmed, Endpoint{
			Method:     "GET",
			Path:       "/",
			Status:     status,
			Confidence: "high",
			Source:     "root",
		})
	}

	paths, edges := discover(base, body)

	r.Topology = append(r.Topology, edges...)
	r.Candidates = append(r.Candidates, paths...)

	for _, p := range smartDefaults(r.Fingerprints, body) {
		r.Candidates = append(r.Candidates, p)
	}

	r.Candidates = uniqStrings(r.Candidates)

	for _, p := range r.Candidates {

		if p == "/" {
			continue
		}

		status, _, _, err := fetch(c, base+p)

		ep := Endpoint{
			Method:     "GET",
			Path:       p,
			Status:     status,
			Confidence: "none",
			Source:     "probe",
		}

		if err != nil {
			r.Rejected = append(r.Rejected, ep)
			continue
		}

		switch {
		case status >= 200 && status < 300:
			ep.Confidence = "high"
			r.Confirmed = append(r.Confirmed, ep)

		case status >= 300 && status < 400:
			ep.Confidence = "medium"
			r.Confirmed = append(r.Confirmed, ep)

		case status == 401 || status == 403 || status == 405:
			ep.Confidence = "medium"
			r.Confirmed = append(r.Confirmed, ep)

		default:
			r.Rejected = append(r.Rejected, ep)
		}
	}

	r.Confirmed = uniqEndpoints(r.Confirmed)
	r.Topology = uniqEdges(r.Topology)

	sort.Slice(r.Confirmed, func(i, j int) bool {
		return r.Confirmed[i].Path < r.Confirmed[j].Path
	})

	if len(r.Confirmed) <= 1 {
		r.Notes = append(r.Notes,
			"No confirmed child endpoints discovered from safe unauthenticated probes.")
	}

	return r, nil
}

func Render(title string, r *ScanResult) string {

	var b strings.Builder

	high := []Endpoint{}
	medium := []Endpoint{}
	denied := []Endpoint{}

	for _, ep := range r.Confirmed {

		switch {

		case ep.Status == 403:
			denied = append(denied, ep)

		case ep.Confidence == "high":
			high = append(high, ep)

		default:
			medium = append(medium, ep)
		}
	}

	fmt.Fprintf(&b, "\033[1;36m%s\033[0m\n", title)
	fmt.Fprintf(&b, "\033[2m%s\033[0m\n\n", r.BaseURL)

	fmt.Fprintf(&b, "Fingerprint\n")
	fmt.Fprintf(&b, "───────────\n")

	fmt.Fprintf(&b, "Type     %s\n", r.APIType)

	if len(r.Fingerprints) > 0 {

		fmt.Fprintf(&b, "Traits   ")

		for i, fp := range r.Fingerprints {

			fp = strings.ReplaceAll(fp, "content-type: ", "")
			fp = strings.ReplaceAll(fp, "server: ", "")

			if i > 0 {
				fmt.Fprintf(&b, " · ")
			}

			fmt.Fprintf(&b, "%s", fp)
		}

		fmt.Fprintf(&b, "\n")
	}

	fmt.Fprintf(&b, "\n")

	fmt.Fprintf(&b, "Discovery\n")
	fmt.Fprintf(&b, "─────────\n")

	if len(high) > 0 {

		fmt.Fprintf(&b, "\033[1;32mLive endpoints\033[0m\n")

		for _, ep := range high {

			fmt.Fprintf(
				&b,
				"  %s  \033[1m%-24s\033[0m  %d\n",
				icon(ep.Path),
				ep.Path,
				ep.Status,
			)
		}

		fmt.Fprintf(&b, "\n")
	}

	if len(medium) > 0 {

		fmt.Fprintf(&b, "\033[1;33mRestricted/discovered\033[0m\n")

		for _, ep := range medium {

			fmt.Fprintf(
				&b,
				"  • %-28s %d\n",
				ep.Path,
				ep.Status,
			)
		}

		fmt.Fprintf(&b, "\n")
	}

	if len(denied) > 0 {

		fmt.Fprintf(&b, "\033[2mDenied probes\033[0m\n")

		for _, ep := range denied {

			fmt.Fprintf(
				&b,
				"  × %s\n",
				ep.Path,
			)
		}

		fmt.Fprintf(&b, "\n")
	}

	fmt.Fprintf(&b, "Topology\n")
	fmt.Fprintf(&b, "────────\n")

	if len(r.Topology) == 0 {

		fmt.Fprintf(&b, "\033[2mNo graph relationships discovered\033[0m\n")

	} else {

		seen := map[string]bool{}

		for _, e := range r.Topology {

			line := fmt.Sprintf(
				"%s → %s [%s]",
				e.From,
				e.To,
				e.Label,
			)

			if seen[line] {
				continue
			}

			seen[line] = true

			fmt.Fprintf(
				&b,
				"  %s\n",
				line,
			)
		}
	}

	if len(r.Notes) > 0 {

		fmt.Fprintf(&b, "\nNotes\n")
		fmt.Fprintf(&b, "─────\n")

		for _, n := range r.Notes {
			fmt.Fprintf(&b, "  %s\n", n)
		}
	}

	return b.String()
}

func icon(path string) string {

	switch {

	case strings.Contains(path, "user"):
		return "👤"

	case strings.Contains(path, "repo"):
		return "📦"

	case strings.Contains(path, "search"):
		return "🔍"

	case strings.Contains(path, "event"):
		return "📡"

	case strings.Contains(path, "rate"):
		return "⏱"

	case strings.Contains(path, "graphql"):
		return "🕸"

	default:
		return "•"
	}
}

func normalize(raw string) (string, error) {

	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "https://")
	raw = strings.TrimPrefix(raw, "http://")
	raw = strings.TrimRight(raw, "/")

	if raw == "" {
		return "", fmt.Errorf("empty target")
	}

	u, err := url.Parse("https://" + raw)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("invalid target")
	}

	return "https://" + u.Host, nil
}

func fetch(c *http.Client, u string) (int, []byte, http.Header, error) {

	req, _ := http.NewRequest("GET", u, nil)

	req.Header.Set("User-Agent", "restless-overlord")
	req.Header.Set("Accept", "application/json,text/html,*/*")

	resp, err := c.Do(req)
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
		out = append(out, "content-type: "+ct)
	}

	if s := h.Get("Server"); s != "" {
		out = append(out, "server: "+s)
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

	low := strings.ToLower(string(body))

	if strings.Contains(low, "graphql") {
		out = append(out, "graphql-hints")
	}

	if strings.Contains(low, "swagger") ||
		strings.Contains(low, "openapi") {

		out = append(out, "openapi-hints")
	}

	return uniqStrings(out)
}

func detectAPIType(fp []string, body []byte) string {

	s := strings.Join(fp, " ")

	switch {

	case strings.Contains(s, "github-api"):
		return "REST catalog"

	case strings.Contains(s, "graphql"):
		return "GraphQL"

	case strings.Contains(s, "json-root"):
		return "REST/JSON"

	case strings.Contains(strings.ToLower(string(body)), "<html"):
		return "web/html"

	default:
		return "unknown"
	}
}

func discover(base string, body []byte) ([]string, []Edge) {

	var paths []string
	var edges []Edge

	var data any

	if json.Unmarshal(body, &data) == nil {
		walkJSON(base, "/", data, &paths, &edges)
	}

	re := regexp.MustCompile(`(?i)(?:href|src|action)=["']([^"']+)["']|fetch\(["']([^"']+)["']\)`)

	for _, m := range re.FindAllStringSubmatch(string(body), -1) {

		for _, g := range m[1:] {

			if p := cleanPath(base, g); p != "" {

				paths = append(paths, p)

				edges = append(edges, Edge{
					From:  "/",
					To:    p,
					Label: "html/js",
				})
			}
		}
	}

	return uniqStrings(paths), uniqEdges(edges)
}

func walkJSON(base, parent string, v any, paths *[]string, edges *[]Edge) {

	switch x := v.(type) {

	case map[string]any:

		for k, v := range x {

			if s, ok := v.(string); ok {

				if p := cleanPath(base, s); p != "" {

					*paths = append(*paths, p)

					*edges = append(*edges, Edge{
						From:  parent,
						To:    p,
						Label: cleanLabel(k),
					})
				}
			}

			walkJSON(base, parent, v, paths, edges)
		}

	case []any:

		for _, v := range x {
			walkJSON(base, parent, v, paths, edges)
		}
	}
}

func cleanPath(base, raw string) string {

	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, `"'`)

	if raw == "" {
		return ""
	}

	baseURL, _ := url.Parse(base)

	u, err := url.Parse(raw)

	if err == nil {

		if u.IsAbs() {

			if baseURL != nil && u.Host != baseURL.Host {
				return ""
			}

			raw = u.Path

		} else if strings.HasPrefix(raw, "/") {

			raw = u.Path
		}
	}

	if strings.Contains(raw, "{") {
		raw = strings.Split(raw, "{")[0]
	}

	raw = strings.TrimSpace(raw)

	if raw == "" {
		return ""
	}

	if !strings.HasPrefix(raw, "/") {
		return ""
	}

	raw = strings.TrimRight(raw, "/")

	if raw == "" {
		return "/"
	}

	return raw
}

func smartDefaults(fp []string, body []byte) []string {

	out := []string{
		"/robots.txt",
		"/sitemap.xml",
		"/openapi.json",
		"/swagger.json",
		"/graphql",
		"/api",
		"/api/v1",
	}

	joined := strings.Join(fp, " ")

	if strings.Contains(joined, "github-api") {

		out = append(out,
			"/user",
			"/users",
			"/users/octocat",
			"/repos",
			"/orgs",
			"/search",
			"/events",
			"/meta",
			"/rate_limit",
			"/zen",
			"/emojis",
		)
	}

	return uniqStrings(out)
}

func cleanLabel(s string) string {

	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "_url")
	s = strings.ReplaceAll(s, "_", "-")

	if s == "" {
		return "link"
	}

	return s
}

func uniqStrings(in []string) []string {

	m := map[string]bool{}
	var out []string

	for _, s := range in {

		s = strings.TrimSpace(s)

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

		k := e.From + "->" + e.To + ":" + e.Label

		if m[k] {
			continue
		}

		m[k] = true
		out = append(out, e)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].To < out[j].To
	})

	return out
}
