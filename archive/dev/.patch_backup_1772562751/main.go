package main

import (
	"fmt"
	"os"

	"github.com/bspippi1337/restless/internal/entry"
)

func printHelp() {
	fmt.Println("restless")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  restless <domain-or-url>              # default: smart")
	fmt.Println("  restless probe <domain-or-url>")
	fmt.Println("  restless smart <domain-or-url>")
	fmt.Println("  restless simulate <domain-or-url>")
	fmt.Println("  restless openapi guard ...            # contract guard (OpenAPI)")
	fmt.Println("  restless openapi diff <old> <new>     # breaking changes + semver hint")
	fmt.Println("  restless <METHOD> <url>               # raw HTTP")
	fmt.Println()
	fmt.Println("OpenAPI guard usage:")
	fmt.Println("  restless openapi guard <METHOD> <pathTemplate> <status> <contentType> <jsonFile> --spec <specRef>")
	fmt.Println("Example:")
	fmt.Println("  restless openapi guard GET /users/{id} 200 application/json ./fixtures/user.json --spec ./openapi.yaml")
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printHelp()
		return
	}

	if args[0] == "--help" || args[0] == "-h" {
		printHelp()
		return
	}

	switch args[0] {
	case "smart", "simulate":
		if err := entry.Smart(args[1:]); err != nil {
			os.Exit(1)
		}
	case "probe":
		if err := entry.Normal(args[1:]); err != nil {
			os.Exit(1)
		}
	case "openapi":
		if err := entry.OpenAPI(args[1:]); err != nil {
			os.Exit(1)
		}
	default:
		if err := entry.Normal(args); err != nil {
			os.Exit(1)
		}
	}
}
