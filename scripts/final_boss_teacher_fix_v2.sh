#!/usr/bin/env bash
set -euo pipefail

FILE="cmd/restless/openapi_cli.go"
VERSION_COMMIT_MSG='fix(openapi): run loads spec from disk via ID (teacher stable)'

echo "================================================="
echo " FINAL BOSS v2: Fix openapi run loader + teacher"
echo "================================================="

mkdir -p "$(dirname "$FILE")"

# If the file is tracked, we can restore it. If not, skip restore.
if git ls-files --error-unmatch "$FILE" >/dev/null 2>&1; then
  echo "==> Restoring tracked file to HEAD"
  git restore "$FILE" 2>/dev/null || git checkout -- "$FILE"
else
  echo "==> File is not tracked by git (unmatched pathspec). Will overwrite cleanly."
fi

echo "==> Writing clean OpenAPI CLI with deterministic disk spec loading"

cat > "$FILE" <<'GO'
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bspippi1337/restless/internal/core/app"
	"github.com/bspippi1337/restless/internal/modules/export"
	"github.com/bspippi1337/restless/internal/modules/openapi"
	"github.com/bspippi1337/restless/internal/modules/session"
	"github.com/bspippi1337/restless/internal/ui/term"
)

func handleOpenAPI(args []string) {
	if len(args) < 1 {
		fmt.Println("usage: openapi <import|ls|endpoints|run>")
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
		if err := openapi.PrintEndpoints(args[1]); err != nil {
			fmt.Println("endpoints error:", err)
			os.Exit(1)
		}

	case "run":
		// run <id> <METHOD> <PATH> [--base URL] [-p k=v]... [-q k=v]... [-H 'K: V']... [-d BODY] [-F @file] [--curl] [--save name] [-set k=v]...
		ra, sessSets, err := parseOpenAPIRunArgs(args[1:])
		if err != nil {
			fmt.Println("run error:", err)
			printOpenAPIRunUsage()
			os.Exit(1)
		}

		// prompt for missing path params (interactive only)
		ra.PathParams, err = promptMissingPathParams(ra.Path, ra.PathParams)
		if err != nil {
			fmt.Println("ERROR:", err)
			os.Exit(1)
		}

		// Build App with session + export (templating + save)
		sess := session.New()
		for k, v := range sessSets {
			sess.Set(k, v)
		}

		mods := []app.Module{
			sess,
			openapi.New(),
			export.New(),
		}
		a, err := app.New(mods)
		if err != nil {
			fmt.Println("ERROR:", err)
			os.Exit(1)
		}

		idx, err := openapi.LoadIndex(ra.ID)
		if err != nil {
			fmt.Println("index error:", err)
			os.Exit(1)
		}

		// IMPORTANT: do NOT trust idx.RawPath (may be URL). Load stored spec from disk by ID.
		openapiDir, err := openapi.DefaultDir()
		if err != nil {
			fmt.Println("ERROR:", err)
			os.Exit(1)
		}
		specPath := filepath.Join(openapiDir, ra.ID+".json")

		spec, err := openapi.LoadSpecFromFile(specPath)
		if err != nil {
			fmt.Println("ERROR: spec:", err)
			os.Exit(1)
		}

		req, curl, err := openapi.BuildRequest(idx, spec, ra)
		if err != nil {
			fmt.Println("ERROR: build:", err)
			os.Exit(1)
		}
		if ra.ShowCurl && curl != "" {
			fmt.Println(curl)
		}

		resp, err := a.RunOnce(context.Background(), req)
		if err != nil {
			fmt.Println("ERROR: request:", err)
			os.Exit(1)
		}

		fmt.Println(term.Status(resp.StatusCode), "(dur=", resp.DurationMs, "ms)")
		fmt.Println(string(resp.Body))

		if ra.SaveAsName != "" {
			p, err := export.SaveJSONArtifact(ra.SaveAsName, resp)
			if err != nil {
				fmt.Println("save error:", err)
				os.Exit(1)
			}
			fmt.Println("saved:", p)
		}

	default:
		fmt.Println("unknown openapi command")
		os.Exit(1)
	}
}

func printOpenAPIRunUsage() {
	fmt.Println("usage:")
	fmt.Println("  openapi run <id> <METHOD> <PATH> [--base URL] [-p k=v]... [-q k=v]... [-H 'K: V']... [-d BODY] [-F @file] [--curl] [--save name] [-set k=v]...")
	fmt.Println("examples:")
	fmt.Println("  restless openapi run <id> GET /pets --base https://petstore3.swagger.io/api/v3")
	fmt.Println("  restless openapi run <id> GET /pets/{petId} --base https://petstore3.swagger.io/api/v3 -p petId=7")
	fmt.Println("  restless openapi run <id> GET /pets --base https://petstore3.swagger.io/api/v3 -q limit=10 --curl")
	fmt.Println("  restless openapi run <id> GET /secure -H 'Authorization: Bearer {{token}}' -set token=abc --base https://example.com")
}

func parseOpenAPIRunArgs(args []string) (openapi.RunArgs, map[string]string, error) {
	if len(args) < 3 {
		return openapi.RunArgs{}, nil, fmt.Errorf("need <id> <method> <path>")
	}
	ra := openapi.RunArgs{
		ID:          args[0],
		Method:      args[1],
		Path:        args[2],
		PathParams:  map[string]string{},
		QueryParams: map[string]string{},
		Headers:     map[string]string{},
	}
	sessSets := map[string]string{}

	i := 3
	for i < len(args) {
		a := args[i]

		switch a {
		case "--base":
			i++
			if i >= len(args) {
				return openapi.RunArgs{}, nil, fmt.Errorf("missing value for --base")
			}
			ra.BaseOverride = args[i]

		case "-p":
			i++
			if i >= len(args) {
				return openapi.RunArgs{}, nil, fmt.Errorf("missing value for -p")
			}
			k, v, ok := splitKV(args[i])
			if !ok {
				return openapi.RunArgs{}, nil, fmt.Errorf("bad -p, want k=v")
			}
			ra.PathParams[k] = v

		case "-q":
			i++
			if i >= len(args) {
				return openapi.RunArgs{}, nil, fmt.Errorf("missing value for -q")
			}
			k, v, ok := splitKV(args[i])
			if !ok {
				return openapi.RunArgs{}, nil, fmt.Errorf("bad -q, want k=v")
			}
			ra.QueryParams[k] = v

		case "-H":
			i++
			if i >= len(args) {
				return openapi.RunArgs{}, nil, fmt.Errorf("missing value for -H")
			}
			k, v, ok := splitHeader(args[i])
			if !ok {
				return openapi.RunArgs{}, nil, fmt.Errorf("bad -H, want 'Key: Value'")
			}
			ra.Headers[k] = v

		case "-d":
			i++
			if i >= len(args) {
				return openapi.RunArgs{}, nil, fmt.Errorf("missing value for -d")
			}
			ra.Body = []byte(args[i])

		case "-F":
			i++
			if i >= len(args) {
				return openapi.RunArgs{}, nil, fmt.Errorf("missing value for -F")
			}
			p := strings.TrimPrefix(args[i], "@")
			b, err := os.ReadFile(p)
			if err != nil {
				return openapi.RunArgs{}, nil, err
			}
			ra.Body = b

		case "--curl":
			ra.ShowCurl = true

		case "--save":
			i++
			if i >= len(args) {
				return openapi.RunArgs{}, nil, fmt.Errorf("missing value for --save")
			}
			ra.SaveAsName = args[i]

		case "-set":
			i++
			if i >= len(args) {
				return openapi.RunArgs{}, nil, fmt.Errorf("missing value for -set")
			}
			k, v, ok := splitKV(args[i])
			if !ok {
				return openapi.RunArgs{}, nil, fmt.Errorf("bad -set, want k=v")
			}
			sessSets[k] = v

		default:
			return openapi.RunArgs{}, nil, fmt.Errorf("unknown arg: %s", a)
		}

		i++
	}

	return ra, sessSets, nil
}

func splitKV(s string) (k, v string, ok bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return s[:i], s[i+1:], true
		}
	}
	return "", "", false
}

func splitHeader(s string) (k, v string, ok bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			k = strings.TrimSpace(s[:i])
			v = strings.TrimSpace(s[i+1:])
			if k == "" {
				return "", "", false
			}
			return k, v, true
		}
	}
	return "", "", false
}

func promptMissingPathParams(path string, params map[string]string) (map[string]string, error) {
	missing := []string{}
	s := path
	for {
		i := strings.Index(s, "{")
		if i == -1 {
			break
		}
		j := strings.Index(s[i:], "}")
		if j == -1 {
			break
		}
		key := s[i+1 : i+j]
		if _, ok := params[key]; !ok && key != "" {
			missing = append(missing, key)
		}
		s = s[i+j+1:]
	}

	if len(missing) == 0 {
		return params, nil
	}

	if !term.IsTTY() {
		return nil, fmt.Errorf("missing path params: %v (non-interactive, pass -p key=value)", missing)
	}

	in := bufio.NewReader(os.Stdin)
	for _, k := range missing {
		fmt.Printf("Enter %s: ", k)
		val, err := in.ReadString('\n')
		if err != nil {
			return nil, err
		}
		val = strings.TrimSpace(val)
		if val == "" {
			return nil, fmt.Errorf("empty value for path param: %s", k)
		}
		params[k] = val
	}
	return params, nil
}
GO

echo "==> gofmt"
go fmt ./...

echo "==> go test"
go test ./...

echo "==> build"
go build -o restless ./cmd/restless

echo "==> teacher"
./restless teacher

echo "==> commit"
git add "$FILE"
git commit -m "$VERSION_COMMIT_MSG" || echo "No changes to commit."

echo "✅ FINAL BOSS v2 DONE"
