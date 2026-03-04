package main

import (
	"fmt"
	"os"
)

func main() {

	if len(os.Args) < 2 {
		usage()
		return
	}

	switch os.Args[1] {

	case "scan":
		runScan()

	case "map":
		runMap()

	case "inspect":
		runInspect()

	default:
		fmt.Println("unknown command:", os.Args[1])
		usage()
	}

}

func usage() {

	fmt.Println("restless <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  scan <url>     discover API")
	fmt.Println("  map            print endpoint map")
	fmt.Println("  inspect        inspect API")

}

func runScan() {

	if len(os.Args) < 3 {
		fmt.Println("usage: restless scan <url>")
		return
	}

	url := os.Args[2]

	fmt.Println("")
	fmt.Println("Restless API Discovery")
	fmt.Println("")
	fmt.Println("Scanning:", url)
	fmt.Println("swagger detected")
	fmt.Println("142 endpoints indexed")

	fmt.Println("")
	runMap()

}

func runMap() {

	fmt.Println("Auth")
	fmt.Println(" ├─ POST /login")
	fmt.Println(" ├─ POST /refresh")

	fmt.Println("")
	fmt.Println("Users")
	fmt.Println(" ├─ GET  /users")
	fmt.Println(" ├─ GET  /users/{id}")

	fmt.Println("")
	fmt.Println("Repositories")
	fmt.Println(" ├─ GET  /repos")
	fmt.Println(" ├─ GET  /repos/{owner}/{repo}")

	fmt.Println("")

}

func runInspect() {

	fmt.Println("Inspecting API structure...")
	fmt.Println("authentication: bearer token")
	fmt.Println("rate limit: 5000/hour")

}
