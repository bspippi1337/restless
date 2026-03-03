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
	fmt.Println("  restless help")
	fmt.Println("  restless --man")
	fmt.Println("  restless <domain-or-url>")
	fmt.Println("  restless smart <domain-or-url>")
	fmt.Println("  restless probe <domain-or-url>")
	fmt.Println("  restless openapi <subcommand>")
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printHelp()
		return
	}

	// global help
	if args[0] == "help" {
		_ = entry.Help(args[1:])
		return
	}

	for _, a := range args {
		if a == "--man" {
			_ = entry.Man(args)
			return
		}
		if a == "--help" || a == "-h" {
			_ = entry.Help(args)
			return
		}
	}

	switch args[0] {

	case "openapi":
		if len(args) == 1 {
			_ = entry.OpenAPI([]string{})
			return
		}
		if err := entry.OpenAPI(args[1:]); err != nil {
			fmt.Println("openapi error:", err)
			os.Exit(1)
		}

	case "smart", "simulate":
		if err := entry.Smart(args[1:]); err != nil {
			os.Exit(1)
		}

	case "probe":
		if err := entry.Normal(args[1:]); err != nil {
			os.Exit(1)
		}

	default:
		if err := entry.Normal(args); err != nil {
			os.Exit(1)
		}
	}
}
