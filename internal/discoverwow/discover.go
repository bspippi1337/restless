package discoverwow

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

type EndpointScore struct {
	Path   string
	Score  int
	Reason string
}

type FieldInfo struct {
	Path   string
	Fields []string
}

type Relation struct {
	From string
	To   []string
}

type Result struct {
	Target    string
	Identity  []string
	Traversal []string
	Schema    []string
	Flows     []string
	Params    []string
	Signals   []string
	Next      []string

	TopEndpoints []EndpointScore
	FieldIntel   []FieldInfo
	Relations    []Relation
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
			path := simplify(v)

			res.Traversal = appendUnique(
				res.Traversal,
				path,
			)

			score := scorePath(path)

			res.TopEndpoints = append(
				res.TopEndpoints,
				EndpointScore{
					Path:   path,
					Score:  score,
					Reason: explainScore(path),
				},
			)

			fields := inspectSample(
				client,
				target,
				path,
			)

			if len(fields) > 0 {
				res.FieldIntel = append(
					res.FieldIntel,
					FieldInfo{
						Path:   path,
						Fields: fields,
					},
				)
			}

			rel := inferRelation(path)

			if rel.From != "" {
				res.Relations = append(
					res.Relations,
					rel,
				)
			}
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

	if resp.Header.Get("ETag") != "" {
		res.Signals = appendUnique(
			res.Signals,
			"conditional request support",
		)
	}

	if resp.Header.Get("X-RateLimit-Limit") != "" {
		res.Signals = appendUnique(
			res.Signals,
			"rate-limit headers",
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

	sort.Slice(
		res.TopEndpoints,
		func(i, j int) bool {
			return res.TopEndpoints[i].Score >
				res.TopEndpoints[j].Score
		},
	)

	return res, nil
}

func Render(r *Result) string {
	var b strings.Builder

	fmt.Fprintf(&b, "\nDISCOVER\n")
	fmt.Fprintf(&b, "Target      %s\n\n", trimProto(r.Target))

	renderSimple(&b, "Identity Model", r.Identity)
	renderEndpoints(&b, r.TopEndpoints)
	renderFields(&b, r.FieldIntel)
	renderRelations(&b, r.Relations)
	renderSimple(&b, "Schema Hints", r.Schema)
	renderSimple(&b, "Flow Candidates", r.Flows)
	renderSimple(&b, "Parameter Inference", r.Params)
	renderSimple(&b, "Interesting Signals", r.Signals)
	renderSimple(&b, "Suggested Next Step", r.Next)

	return b.String()
}

func renderSimple(
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

func renderEndpoints(
	b *strings.Builder,
	items []EndpointScore,
) {
	if len(items) == 0 {
		return
	}

	fmt.Fprintf(b, "Top Candidates\n")
	fmt.Fprintf(b, "--------------\n")

	limit := 8

	if len(items) < limit {
		limit = len(items)
	}

	for i := 0; i < limit; i++ {
		it := items[i]

		fmt.Fprintf(
			b,
			"  %2d  %-50s %s\n",
			it.Score,
			it.Path,
			it.Reason,
		)
	}

	fmt.Fprintln(b)
}

func renderFields(
	b *strings.Builder,
	items []FieldInfo,
) {
	if len(items) == 0 {
		return
	}

	fmt.Fprintf(b, "Field Intelligence\n")
	fmt.Fprintf(b, "------------------\n")

	limit := 3

	if len(items) < limit {
		limit = len(items)
	}

	for i := 0; i < limit; i++ {
		it := items[i]

		fmt.Fprintf(b, "  %s\n", it.Path)

		for _, f := range it.Fields {
			fmt.Fprintf(b, "    - %s\n", f)
		}

		fmt.Fprintln(b)
	}
}

func renderRelations(
	b *strings.Builder,
	items []Relation,
) {
	if len(items) == 0 {
		return
	}

	fmt.Fprintf(b, "Relationship Graph\n")
	fmt.Fprintf(b, "------------------\n")

	for _, r := range items {
		fmt.Fprintf(b, "  %s\n", r.From)

		for _, to := range r.To {
			fmt.Fprintf(b, "    -> %s\n", to)
		}

		fmt.Fprintln(b)
	}
}

func inspectSample(
	client *http.Client,
	target string,
	path string,
) []string {
	if strings.Contains(path, "{") {
		return nil
	}

	req, _ := http.NewRequest(
		"GET",
		target+path,
		nil,
	)

	req.Header.Set(
		"User-Agent",
		"restless-discover/next",
	)

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil
	}

	body, _ := io.ReadAll(
		io.LimitReader(resp.Body, 8192),
	)

	var obj map[string]interface{}

	if json.Unmarshal(body, &obj) != nil {
		return nil
	}

	fields := make([]string, 0)

	for k := range obj {
		fields = append(fields, k)
	}

	sort.Strings(fields)

	if len(fields) > 6 {
		fields = fields[:6]
	}

	return fields
}

func inferRelation(path string) Relation {
	switch {
	case strings.Contains(path, "user"):
		return Relation{
			From: "user",
			To: []string{
				"repositories",
				"followers",
				"organizations",
			},
		}

	case strings.Contains(path, "repo"):
		return Relation{
			From: "repository",
			To: []string{
				"issues",
				"contributors",
				"commits",
			},
		}
	}

	return Relation{}
}

func scorePath(path string) int {
	score := 50

	switch {
	case strings.Contains(path, "search"):
		score += 40
	case strings.Contains(path, "repo"):
		score += 38
	case strings.Contains(path, "user"):
		score += 35
	case strings.Contains(path, "event"):
		score += 30
	}

	if strings.Contains(path, "{") {
		score += 8
	}

	return score
}

func explainScore(path string) string {
	switch {
	case strings.Contains(path, "search"):
		return "query surface"
	case strings.Contains(path, "repo"):
		return "repository traversal"
	case strings.Contains(path, "user"):
		return "identity traversal"
	case strings.Contains(path, "event"):
		return "activity stream"
	}

	return "general endpoint"
}

func inferParams(s string) []string {
	var out []string

	re := regexp.MustCompile(
		`[?&]([a-zA-Z0-9_\-]+)=`,
	)

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

func hasRelation(in []Relation, r Relation) bool {
	for _, x := range in {
		if x.From != r.From {
			continue
		}

		if len(x.To) != len(r.To) {
			continue
		}

		match := true

		for i := range x.To {
			if x.To[i] != r.To[i] {
				match = false
				break
			}
		}

		if match {
			return true
		}
	}

	return false
}

func trimProto(s string) string {
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	return s
}
