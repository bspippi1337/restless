package engine

import "fmt"

func PrintResult(r *Result) {

	fmt.Println("Fingerprint")
	fmt.Println("-----------")
	fmt.Println("API type:", r.APIType)
	fmt.Println()

	fmt.Println("Endpoints discovered")
	fmt.Println("--------------------")

	for _, e := range r.Endpoints {
		fmt.Printf("[%s] %s\n", e.Confidence, e.Path)
	}

	fmt.Println()
	fmt.Println("Topology")
	fmt.Println("--------")
	fmt.Println(r.Topology)
}
