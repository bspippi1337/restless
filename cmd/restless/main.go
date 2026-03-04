package main

import (
	"fmt"
	"os"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("restless <command>")
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  scan <url>     discover API")
		fmt.Println("  map            print endpoint map")
		fmt.Println("  inspect        inspect API")
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
	}

}

func runScan() {
	fmt.Println("scan: not implemented yet (HN demo stub)")
}

func runMap() {
	fmt.Println("map: not implemented yet (HN demo stub)")
}

func runInspect() {
	fmt.Println("inspect: not implemented yet (HN demo stub)")
}
