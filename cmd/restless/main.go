package main

import (
	"fmt"
	"os"

	"github.com/bspippi1337/restless/internal/discovery"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage: restless inspect <target>")
		os.Exit(1)
	}

	cmd := os.Args[1]
	target := os.Args[2]

	if cmd != "inspect" {
		fmt.Println("currently only inspect is patched in this recovery build")
		os.Exit(1)
	}

	fp, err := discovery.FingerprintTarget(target)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("Restless Real Discovery Engine")
	fmt.Println()

	fmt.Printf("Scanning: %s\n\n", fp.Target)

	fmt.Println("[1/5] probing target")
	fmt.Println("[2/5] fingerprinting stack")
	fmt.Println("[3/5] extracting API hints")
	fmt.Println("[4/5] inferring architecture")
	fmt.Println("[5/5] scoring confidence")

	fmt.Println()
	fmt.Println("Fingerprint")
	fmt.Println("-----------")
	fmt.Printf("API type: %s\n", fp.APIType)
	fmt.Printf("Server: %s\n", fp.Server)
	fmt.Printf("Confidence: %d/100\n", fp.Confidence)

	fmt.Println()
	fmt.Println("Technologies")
	fmt.Println("------------")

	if len(fp.Technologies) == 0 {
		fmt.Println("No obvious technologies detected")
	} else {
		for _, t := range fp.Technologies {
			fmt.Printf("- %s\n", t)
		}
	}

	fmt.Println()
	fmt.Println("Discovery")
	fmt.Println("---------")

	if fp.GraphQL {
		fmt.Println("- GraphQL hints detected")
	}

	if fp.OpenAPI {
		fmt.Println("- OpenAPI/Swagger hints detected")
	}

	if len(fp.InterestingURLs) == 0 {
		fmt.Println("- No obvious API endpoints discovered")
	} else {
		for _, u := range fp.InterestingURLs {
			fmt.Printf("- %s\n", u)
		}
	}
}
