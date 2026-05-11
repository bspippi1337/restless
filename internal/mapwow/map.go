package mapwow

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

type Node struct {
	Name  string
	Type  string
	Edges []string
}

type Result struct {
	Target   string
	Nodes    []Node
	Signals  []string
	Topology []string
}

func Map(target string) (*Result, error) {
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
		Target: target,
	}

	body, _ := io.ReadAll(
		io.LimitReader(resp.Body, 65536),
	)

	var obj map[string]interface{}

	if json.Unmarshal(body, &obj) == nil {
		keys := make([]string, 0)

		for k := range obj {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, k := range keys {
			v := obj[k]

			node := Node{
				Name: k,
				Type: detectType(v),
			}

			switch vv := v.(type) {
			case map[string]interface{}:
				for sub := range vv {
					node.Edges = append(
						node.Edges,
						sub,
					)
				}

			case []interface{}:
				node.Edges = append(
					node.Edges,
					"collection",
				)
			}

			r.Nodes = append(
				r.Nodes,
				node,
			)
		}
	}

	r.Topology = inferTopology(r.Nodes)
	r.Signals = inferSignals(r.Nodes)

	if strings.Contains(
		strings.ToLower(target),
		"github.com",
	) {
		enrichGitHub(r)

		for _, t := range inferGitHubTopology(r.Nodes) {
			r.Topology = appendUnique(
				r.Topology,
				t,
			)
		}
	}

	return r, nil
}

func Render(r *Result) string {
	var b strings.Builder

	fmt.Fprintf(&b, "\nMAP\n")
	fmt.Fprintf(&b, "Target  %s\n\n", trimProto(r.Target))

	if len(r.Signals) > 0 {
		fmt.Fprintf(&b, "Surface Signals\n")
		fmt.Fprintf(&b, "---------------\n")

		for _, s := range r.Signals {
			fmt.Fprintf(&b, "- %s\n", s)
		}

		fmt.Fprintln(&b)
	}

	if len(r.Nodes) > 0 {
		fmt.Fprintf(&b, "Entity Graph\n")
		fmt.Fprintf(&b, "------------\n")

		limit := 20

		if len(r.Nodes) < limit {
			limit = len(r.Nodes)
		}

		for i := 0; i < limit; i++ {
			n := r.Nodes[i]

			fmt.Fprintf(
				&b,
				"%s [%s]\n",
				n.Name,
				n.Type,
			)

			for _, e := range n.Edges {
				fmt.Fprintf(
					&b,
					"  -> %s\n",
					e,
				)
			}

			fmt.Fprintln(&b)
		}
	}

	if len(r.Topology) > 0 {
		fmt.Fprintf(&b, "Topology\n")
		fmt.Fprintf(&b, "--------\n")

		for _, t := range r.Topology {
			fmt.Fprintf(&b, "- %s\n", t)
		}
	}

	return b.String()
}

func inferTopology(nodes []Node) []string {
	var out []string

	for _, n := range nodes {
		l := strings.ToLower(n.Name)

		switch {
		case strings.Contains(l, "user"):
			out = appendUnique(
				out,
				"user-centric graph",
			)

		case strings.Contains(l, "repo"):
			out = appendUnique(
				out,
				"repository-linked graph",
			)

		case strings.Contains(l, "url"):
			out = appendUnique(
				out,
				"hypermedia traversal",
			)
		}
	}

	return out
}

func inferSignals(nodes []Node) []string {
	var out []string

	for _, n := range nodes {
		switch n.Type {
		case "array":
			out = appendUnique(
				out,
				"collection-oriented schema",
			)

		case "object":
			out = appendUnique(
				out,
				"nested entity graph",
			)
		}
	}

	return out
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
