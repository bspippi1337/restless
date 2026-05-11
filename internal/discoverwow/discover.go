package discoverwow

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Result struct {
	Target    string
	Identity  []string
	Traversal []string
	Schema    []string
	Flows     []string
	Params    []string
	Signals   []string
	Next      []string
}

func Discover(target string) (*Result, error) {
	if !strings.HasPrefix(target, "http") {
		target = "https://" + target
	}

	client := &http.Client{
		Timeout: 7 * time.Second,
	}

	res := &Result{
		Target: target,
	}

	req, _ := http.NewRequest("GET", target, nil)
	req.Header.Set("User-Agent", "restless-discover/next")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var root map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&root)

	keys := make([]string, 0, len(root))

	for k := range root {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	paramSeen := map[string]bool{}

	for _, k := range keys {
		v := fmt.Sprintf("%v", root[k])

		if looksURL(v) {
			res.Traversal = appendUnique(
				res.Traversal,
				simplify(v),
			)
		}

		lk := strings.ToLower(k)

		switch {
		case strings.Contains(lk, "user"):
			res.Identity = appendUnique(
				res.Identity,
				"user-centric namespace",
			)

		case strings.Contains(lk, "org"):
			res.Identity = appendUnique(
				res.Identity,
				"organization hierarchy",
			)

		case strings.Contains(lk, "repo"):
			res.Signals = appendUnique(
				res.Signals,
				"repository-oriented platform",
			)

		case strings.Contains(lk, "search"):
			res.Signals = appendUnique(
				res.Signals,
				"indexed search surface",
			)

		case strings.Contains(lk, "event"):
			res.Signals = appendUnique(
				res.Signals,
				"public event stream",
			)
		}

		for _, p := range inferParams(v) {
			if !paramSeen[p] {
				paramSeen[p] = true
				res.Params = append(res.Params, p)
			}
		}
	}

	link := resp.Header.Get("Link")

	if strings.Contains(link, `rel="next"`) {
		res.Schema = append(
			res.Schema,
			"pagination detected",
		)
	}

	if resp.Header.Get("X-RateLimit-Limit") != "" {
		res.Signals = append(
			res.Signals,
			"rate-limit headers",
		)
	}

	if resp.Header.Get("ETag") != "" {
		res.Signals = append(
			res.Signals,
			"conditional request support",
		)
	}

	if len(res.Traversal) > 0 {
		res.Flows = append(
			res.Flows,
			"enumerate resources",
			"inspect entities",
			"traverse relations",
		)

		res.Next = append(
			res.Next,
			"restless inspect "+target+res.Traversal[0],
		)
	}

	sort.Strings(res.Traversal)
	sort.Strings(res.Params)
	sort.Strings(res.Signals)

	return res, nil
}

func Render(r *Result) string {
	var b strings.Builder

	fmt.Fprintf(&b, "\nDISCOVER\n")
	fmt.Fprintf(&b, "Target      %s\n\n", trimProto(r.Target))

	renderSection(&b, "Identity Model", r.Identity)
	renderSection(&b, "Traversal Candidates", r.Traversal)
	renderSection(&b, "Schema Hints", r.Schema)
	renderSection(&b, "Flow Candidates", r.Flows)
	renderSection(&b, "Parameter Inference", r.Params)
	renderSection(&b, "Interesting Signals", r.Signals)
	renderSection(&b, "Suggested Next Step", r.Next)

	return b.String()
}

func renderSection(
	b *strings.Builder,
	title string,
	items []string,
) {
	if len(items) == 0 {
		return
	}

	fmt.Fprintf(b, "%s\n", title)
	fmt.Fprintf(b, "%s\n", strings.Repeat("-", len(title)))

	for _, item := range items {
		fmt.Fprintf(b, "  - %s\n", item)
	}

	fmt.Fprintln(b)
}

func inferParams(s string) []string {
	var out []string

	re := regexp.MustCompile(`[?&]([a-zA-Z0-9_\-]+)=`)
	matches := re.FindAllStringSubmatch(s, -1)

	for _, m := range matches {
		if len(m) > 1 {
			out = appendUnique(out, m[1]+"=")
		}
	}

	return out
}

func appendUnique(in []string, v string) []string {
	for _, x := range in {
		if x == v {
			return in
		}
	}

	return append(in, v)
}

func looksURL(s string) bool {
	return strings.HasPrefix(s, "http://") ||
		strings.HasPrefix(s, "https://")
}

func simplify(s string) string {
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")

	idx := strings.Index(s, "/")

	if idx == -1 {
		return "/"
	}

	return s[idx:]
}

func trimProto(s string) string {
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	return s
}
