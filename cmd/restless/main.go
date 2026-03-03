package main

import (
	"fmt"
	"os"
	"os/exec"
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
		cmd := exec.Command(os.Args[0], append([]string{"verify"}, os.Args[2:]...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()

	case "map":
		runMap(os.Args[2:])

	case "inspect":
		runInspect(os.Args[2:])

	default:
		fmt.Println("unknown command:", os.Args[1])
	}
}
