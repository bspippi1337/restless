package inspectwow

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

type Field struct {
	Name string
	Type string
}

type Result struct {
	Target      string
	Status      int
	ContentType string

	Fields    []Field
	Relations []string
	Signals   []string
	Examples  []string
	Problems  []string
}

func Inspect(target string) (*Result, error) {
	if !strings.HasPrefix(target, "http") {
		target = "https://" + target
	}

	client := &http.Client{
		Timeout: 8 * time.Second,
	}

	req, _ := http.NewRequest(
		"GET",
		target,
		nil,
	)

	req.Header.Set(
		"User-Agent",
		"Restless/420",
	)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r := &Result{
		Target:      target,
		Status:      resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
	}

	if resp.Header.Get("ETag") != "" {
		r.Signals = append(
			r.Signals,
			"etag support",
		)
	}

	if resp.Header.Get("X-RateLimit-Limit") != "" {
		r.Signals = append(
			r.Signals,
			"rate limited",
		)
	}

	body, _ := io.ReadAll(
		io.LimitReader(resp.Body, 32768),
	)

	if strings.Contains(
		strings.ToLower(target),
		"github.com",
	) {
		enrichGitHub(
			req,
			resp,
			r,
			body,
		)
	}

	var obj map[string]interface{}

	if json.Unmarshal(body, &obj) == nil {

		if msg, ok := obj["message"].(string); ok {
			switch {
			case strings.Contains(
				strings.ToLower(msg),
				"rate limit",
			):
				r.Problems = append(
					r.Problems,
					"github rate limit exceeded",
				)

			case resp.StatusCode == 403:
				r.Problems = append(
					r.Problems,
					"access restricted",
				)
			}
		}

		keys := make([]string, 0)

		for k, v := range obj {
			r.Fields = append(
				r.Fields,
				Field{
					Name: k,
					Type: detectType(v),
				},
			)

			keys = append(keys, k)
		}

		sort.Slice(
			r.Fields,
			func(i, j int) bool {
				return r.Fields[i].Name <
					r.Fields[j].Name
			},
		)

		for _, k := range keys {
			lk := strings.ToLower(k)

			switch {
			case strings.Contains(lk, "user"):
				r.Relations = appendUnique(
					r.Relations,
					"user",
				)

			case strings.Contains(lk, "repo"):
				r.Relations = appendUnique(
					r.Relations,
					"repository",
				)

			case strings.Contains(lk, "issue"):
				r.Relations = appendUnique(
					r.Relations,
					"issues",
				)

			case strings.Contains(lk, "commit"):
				r.Relations = appendUnique(
					r.Relations,
					"commits",
				)
			}
		}
	}

	r.Examples = append(
		r.Examples,
		fmt.Sprintf("GET %s", target),
	)

	return r, nil
}

func Render(r *Result) string {
	var b strings.Builder

	fmt.Fprintf(&b, "\nINSPECT\n")
	fmt.Fprintf(&b, "Target  %s\n", trimProto(r.Target))
	fmt.Fprintf(&b, "Status  %d\n", r.Status)
	fmt.Fprintf(&b, "Content  %s\n\n", r.ContentType)

	if len(r.Problems) > 0 {
		fmt.Fprintf(&b, "Observations\n")
		fmt.Fprintf(&b, "------------\n")

		for _, p := range r.Problems {
			fmt.Fprintf(&b, "- %s\n", p)
		}

		fmt.Fprintln(&b)
	}

	if len(r.Fields) > 0 {
		fmt.Fprintf(&b, "Fields\n")
		fmt.Fprintf(&b, "------\n")

		limit := 16

		if len(r.Fields) < limit {
			limit = len(r.Fields)
		}

		for i := 0; i < limit; i++ {
			f := r.Fields[i]

			fmt.Fprintf(
				&b,
				"%-28s %s\n",
				f.Name,
				f.Type,
			)
		}

		fmt.Fprintln(&b)
	}

	if len(r.Relations) > 0 {
		fmt.Fprintf(&b, "Relations\n")
		fmt.Fprintf(&b, "---------\n")

		for _, rel := range r.Relations {
			fmt.Fprintf(&b, "- %s\n", rel)
		}

		fmt.Fprintln(&b)
	}

	if len(r.Signals) > 0 {
		fmt.Fprintf(&b, "Protocol\n")
		fmt.Fprintf(&b, "--------\n")

		for _, s := range r.Signals {
			fmt.Fprintf(&b, "- %s\n", s)
		}

		fmt.Fprintln(&b)
	}

	if len(r.Examples) > 0 {
		fmt.Fprintf(&b, "Examples\n")
		fmt.Fprintf(&b, "--------\n")

		for _, e := range r.Examples {
			fmt.Fprintf(&b, "%s\n", e)
		}
	}

	return b.String()
}

func detectType(v interface{}) string {
	switch v.(type) {
	case string:
		return "string"
	case bool:
		return "boolean"
	case float64:
		return "number"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

func appendUnique(in []string, v string) []string {
	for _, x := range in {
		if x == v {
			return in
		}
	}

	return append(in, v)
}

func trimProto(s string) string {
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	return s
}
