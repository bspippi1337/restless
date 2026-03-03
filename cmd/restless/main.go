package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {

	if len(os.Args) < 2 {
		fmt.Println("restless")
		fmt.Println("commands: scan verify map inspect version")
		return
	}

	switch os.Args[1] {

	case "version":
		fmt.Println(version)

	case "scan":
		runScan(os.Args[2:])

	case "verify":
		runVerify(os.Args[2:])

	case "map":
		runMap(os.Args[2:])

	case "inspect":
		runInspect(os.Args[2:])

	default:
		fmt.Println("unknown command:", os.Args[1])
	}
}
