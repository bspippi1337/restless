#!/usr/bin/env bash
set -euo pipefail

echo "==> Upgrading OpenAPI to real endpoint-aware mode"

mkdir -p internal/modules/openapi

# --- Spec parsing (minimal OAS3 support) ---

cat > internal/modules/openapi/spec.go <<'EOT'
package openapi

import (
	"encoding/json"
	"errors"
	"os"
)

type Spec struct {
	OpenAPI string                 `json:"openapi"`
	Info    map[string]any         `json:"info"`
	Paths   map[string]PathItem    `json:"paths"`
}

type PathItem map[string]Operation

type Operation struct {
	Summary     string `json:"summary"`
	OperationID string `json:"operationId"`
}

func LoadSpecFromFile(path string) (Spec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Spec{}, err
	}

	var s Spec
	if err := json.Unmarshal(b, &s); err != nil {
		return Spec{}, err
	}

	if s.Paths == nil {
		return Spec{}, errors.New("invalid spec: no paths")
	}

	return s, nil
}
EOT

# --- CLI integration ---

cat > internal/modules/openapi/commands.go <<'EOT'
package openapi

import (
	"fmt"
	"os"
	"path/filepath"
)

func ListSpecs() error {
	dir, err := cacheDir()
	if err != nil {
		return err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			fmt.Println(f.Name())
		}
	}
	return nil
}

func ListEndpoints(id string) error {
	idx, err := LoadIndex(id)
	if err != nil {
		return err
	}

	spec, err := LoadSpecFromFile(idx.RawPath)
	if err != nil {
		return err
	}

	for path, ops := range spec.Paths {
		for method := range ops {
			fmt.Printf("%s %s\n", method, path)
		}
	}
	return nil
}
EOT

# --- Wire into CLI ---

cat > cmd/restless-v2/openapi_cli.go <<'EOT'
package main

import (
	"fmt"
	"os"

	"github.com/bspippi1337/restless/internal/modules/openapi"
)

func handleOpenAPI(args []string) {
	if len(args) < 1 {
		fmt.Println("usage: openapi <import|ls|endpoints>")
		os.Exit(1)
	}

	switch args[0] {
	case "import":
		if len(args) < 2 {
			fmt.Println("usage: openapi import <url|file>")
			os.Exit(1)
		}
		idx, err := openapi.Import(args[1])
		if err != nil {
			fmt.Println("import error:", err)
			os.Exit(1)
		}
		fmt.Println("imported:", idx.ID)

	case "ls":
		if err := openapi.ListSpecs(); err != nil {
			fmt.Println("ls error:", err)
			os.Exit(1)
		}

	case "endpoints":
		if len(args) < 2 {
			fmt.Println("usage: openapi endpoints <id>")
			os.Exit(1)
		}
		if err := openapi.ListEndpoints(args[1]); err != nil {
			fmt.Println("endpoints error:", err)
			os.Exit(1)
		}

	default:
		fmt.Println("unknown openapi command")
		os.Exit(1)
	}
}
EOT

# --- Modify main.go to dispatch ---

cat > cmd/restless-v2/main.go <<'EOT'
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bspippi1337/restless/internal/core/app"
	"github.com/bspippi1337/restless/internal/core/types"
	"github.com/bspippi1337/restless/internal/modules/bench"
	"github.com/bspippi1337/restless/internal/modules/export"
	"github.com/bspippi1337/restless/internal/modules/openapi"
	"github.com/bspippi1337/restless/internal/modules/session"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "openapi" {
		handleOpenAPI(os.Args[2:])
		return
	}

	var (
		method = flag.String("X", "GET", "HTTP method")
		url    = flag.String("url", "", "Request URL")
		body   = flag.String("d", "", "Body string")
	)
	flag.Parse()

	if *url == "" {
		fmt.Println("missing -url")
		os.Exit(1)
	}

	sess := session.New()
	mods := []app.Module{
		sess,
		openapi.New(),
		export.New(),
		bench.New(),
	}

	a, err := app.New(mods)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	req := types.Request{
		Method:  *method,
		URL:     *url,
		Headers: http.Header{},
		Body:    []byte(*body),
	}

	resp, err := a.RunOnce(context.Background(), req)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Printf("status: %d\n", resp.StatusCode)
	fmt.Println(string(resp.Body))
}
EOT

gofmt -w .
go test ./...

echo "OpenAPI v1 applied."
