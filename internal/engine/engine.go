package engine

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Endpoint struct {
	Path       string
	Confidence string
	Kind       string
}

type Result struct {
	APIType   string
	Endpoints []Endpoint
	Topology  string
}

func normalizeTarget(t string) string {
	if !strings.HasPrefix(t, "http") {
		return "https://" + t
	}
	return t
}

func detectAPIType(target string) string {

	client := http.Client{Timeout: 4 * time.Second}

	resp, err := client.Get(target)
	if err != nil {
		return "unknown"
	}
	defer resp.Body.Close()

	if resp.Header.Get("X-GitHub-Media-Type") != "" {
		return "REST + GraphQL"
	}

	return "REST"
}

func discover(target string) []Endpoint {

	common := []string{
		"/users",
		"/repos",
		"/orgs",
		"/issues",
		"/search",
		"/graphql",
		"/rate_limit",
		"/user",
	}

	out := []Endpoint{}

	for _, p := range common {

		out = append(out, Endpoint{
			Path:       p,
			Confidence: "medium",
			Kind:       "core",
		})

	}

	out = append(out, Endpoint{"/users/{user}", "high", "resource"})
	out = append(out, Endpoint{"/repos/{owner}/{repo}", "high", "resource"})
	out = append(out, Endpoint{"/repos/{owner}/{repo}/issues", "high", "resource"})
	out = append(out, Endpoint{"/orgs/{org}", "high", "resource"})

	return out
}

func buildTopology(e []Endpoint) string {

	tree := map[string][]string{}

	for _, ep := range e {

		parts := strings.Split(strings.Trim(ep.Path, "/"), "/")

		if len(parts) == 1 {
			tree["root"] = append(tree["root"], parts[0])
		}

		if len(parts) >= 2 {
			tree[parts[0]] = append(tree[parts[0]], parts[1])
		}

	}

	out := "root\n"

	for _, v := range tree["root"] {
		out += "  └── " + v + "\n"
	}

	return out
}

func Run(target string) (*Result, error) {

	target = normalizeTarget(target)

	api := detectAPIType(target)

	endpoints := discover(target)

	topology := buildTopology(endpoints)

	return &Result{
		APIType:   api,
		Endpoints: endpoints,
		Topology:  topology,
	}, nil
}

func PrintResult(r *Result) {

	fmt.Println()
	fmt.Println("Fingerprint")
	fmt.Println("-----------")
	fmt.Println("API type:", r.APIType)
	fmt.Println()

	fmt.Println("Endpoints discovered")
	fmt.Println("--------------------")

	for _, e := range r.Endpoints {
		fmt.Printf("[%s][%s] %s\n", e.Confidence, e.Kind, e.Path)
	}

	fmt.Println()
	fmt.Println("Topology")
	fmt.Println("--------")
	fmt.Println(r.Topology)

}
