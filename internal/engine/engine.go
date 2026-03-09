package engine

import "fmt"

type Result struct {
	APIType   string
	Endpoints []Endpoint
	Topology  string
	Workflow  []string
}

func Run(target string) (*Result, error) {
	target = normalizeTarget(target)

	apiType := DetectAPIType(target)
	endpoints := DiscoverEndpoints(target)

	var paths []string
	for i := range endpoints {
		endpoints[i].Type = classifyEndpoint(endpoints[i].Path)
		paths = append(paths, endpoints[i].Path)
	}

	topology := BuildTopology(paths)
	workflow := SuggestWorkflow(apiType, target)

	return &Result{
		APIType:   apiType,
		Endpoints: endpoints,
		Topology:  topology,
		Workflow:  workflow,
	}, nil
}

func Print(r *Result) {
	fmt.Println("Fingerprint")
	fmt.Println("-----------")
	fmt.Println("API type:", r.APIType)

	fmt.Println()
	fmt.Println("Endpoints discovered")
	fmt.Println("--------------------")
	for _, e := range r.Endpoints {
		fmt.Printf("[%s][%s] %s\n", e.Confidence, e.Type, e.Path)
	}

	fmt.Println()
	fmt.Println("Topology")
	fmt.Println("--------")
	fmt.Println(r.Topology)

	fmt.Println()
	fmt.Println("Suggested workflows")
	fmt.Println("-------------------")
	for _, w := range r.Workflow {
		fmt.Println(w)
	}
}
