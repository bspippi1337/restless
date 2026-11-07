#!/usr/bin/env bash

cd "$(dirname "$0")/.."

echo "== installing target normalization =="

cat > internal/engine/normalize.go <<'GO'
package engine

import (
	"net/url"
	"strings"
)

func NormalizeTarget(t string) string {

	if !strings.HasPrefix(t, "http") {
		t = "https://" + t
	}

	u, err := url.Parse(t)
	if err != nil {
		return t
	}

	u.Path = ""
	u.RawQuery = ""

	return u.String()
}
GO


echo "== installing pipeline progress =="

cat > internal/engine/pipeline.go <<'GO'
package engine

import "fmt"

func Step(n int, total int, name string) {
	fmt.Printf("[%d/%d] %s\n", n, total, name)
}
GO


echo "== upgrading autopilot engine =="

cat > cmd/restless/main.go <<'GO'
package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/bspippi1337/restless/internal/cli"
	"github.com/bspippi1337/restless/internal/engine"
)

func looksLikeTarget(s string) bool {
	return strings.Contains(s, ".")
}

func openFile(path string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}

	cmd.Start()
}

func main() {

	open := false
	target := ""

	for _, a := range os.Args[1:] {

		if a == "--open" || a == "-o" {
			open = true
			continue
		}

		if looksLikeTarget(a) {
			target = a
		}
	}

	if target != "" {

		target = engine.NormalizeTarget(target)

		fmt.Println("Restless API Discovery Engine")
		fmt.Println("Scanning:", target)
		fmt.Println()

		engine.Step(1,5,"probing API surface")

		res, err := engine.Run(target)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		engine.Step(2,5,"inferring resource model")

		engine.Step(3,5,"building topology")

		engine.PrintResult(res)

		engine.Step(4,5,"generating graph")

		dot := engine.TopologyToDOT(res.Topology)

		out := strings.ReplaceAll(target,"https://","") + ".svg"

		err = engine.RenderDOT(dot,out)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		engine.Step(5,5,"complete")

		fmt.Println()
		fmt.Println("Graph written to",out)

		if open {
			openFile(out)
		}

		os.Exit(0)
	}

	cli.Execute()
}
GO


echo "== adding summary printer =="

cat > internal/engine/summary.go <<'GO'
package engine

import "fmt"

func Summary(r *Result) {

	fmt.Println()
	fmt.Println("API Summary")
	fmt.Println("-----------")

	fmt.Printf("endpoints discovered: %d\n",len(r.Endpoints))

	resources := 0

	for _,e := range r.Endpoints {
		if e.Confidence == "high" {
			resources++
		}
	}

	fmt.Printf("resource endpoints: %d\n",resources)
}
GO


echo "== patching printer =="

cat > internal/engine/print.go <<'GO'
package engine

import "fmt"

func PrintResult(r *Result) {

	fmt.Println()
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

	Summary(r)
}
GO


echo "== formatting =="
go fmt ./...

echo "== rebuilding =="
rm -rf build
make build

echo
echo "===================================="
echo "Restless upgraded"
echo "===================================="
echo
echo "Run:"
echo "./build/restless api.github.com"
