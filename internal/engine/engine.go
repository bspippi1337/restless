package engine

import (
	"fmt"
	"strings"

	"github.com/bspippi1337/restless/internal/discovery"
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

func buildTopology(e []Endpoint) string {
	if len(e) == 0 {
		return "No topology discovered"
	}

	var b strings.Builder
	b.WriteString("root\n")

	for _, ep := range e {
		b.WriteString("  └── ")
		b.WriteString(strings.TrimPrefix(ep.Path, "/"))
		b.WriteString("\n")
	}

	return b.String()
}

func Run(target string) (*Result, error) {
	fp, err := discovery.FingerprintTarget(target)
	if err != nil {
		return nil, err
	}

	endpoints := []Endpoint{}

	for _, u := range fp.InterestingURLs {
		endpoints = append(endpoints, Endpoint{
			Path:       u,
			Confidence: "observed",
			Kind:       "discovered",
		})
	}

	if fp.GraphQL {
		endpoints = append(endpoints, Endpoint{
			Path:       "/graphql",
			Confidence: "high",
			Kind:       "graphql",
		})
	}

	if fp.OpenAPI {
		endpoints = append(endpoints, Endpoint{
			Path:       "/swagger",
			Confidence: "medium",
			Kind:       "openapi",
		})
	}

	return &Result{
		APIType:   fp.APIType,
		Endpoints: endpoints,
		Topology:  buildTopology(endpoints),
	}, nil
}

func PrintResult(r *Result) {
	fmt.Println()
	fmt.Println("Fingerprint")
	fmt.Println("-----------")
	fmt.Println("API type:", r.APIType)
	fmt.Println()

	fmt.Println("Discovery")
	fmt.Println("---------")

	if len(r.Endpoints) == 0 {
		fmt.Println("No obvious API endpoints discovered")
	} else {
		for _, e := range r.Endpoints {
			fmt.Printf("[%s][%s] %s\n", e.Confidence, e.Kind, e.Path)
		}
	}

	fmt.Println()
	fmt.Println("Topology")
	fmt.Println("--------")
	fmt.Println(r.Topology)
}
