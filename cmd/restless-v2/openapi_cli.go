package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bspippi1337/restless/internal/core/app"
	"github.com/bspippi1337/restless/internal/modules/export"
	"github.com/bspippi1337/restless/internal/modules/openapi"
	"github.com/bspippi1337/restless/internal/modules/session"
	"github.com/bspippi1337/restless/internal/profile"
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

		ra, sessSets, err := parseOpenAPIRunArgs(args[1:])
		if err != nil {
			fmt.Println("run error:", err)
			printOpenAPIRunUsage()
			os.Exit(1)
		}

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
			fmt.Println("error:", err)
			os.Exit(1)
		}

		idx, err := openapi.LoadIndex(ra.ID)
		if err != nil {
			fmt.Println("index error:", err)
			os.Exit(1)
		}

		spec, err := openapi.LoadSpecFromFile(idx.RawPath)
		if err != nil {
			fmt.Println("spec error:", err)
			os.Exit(1)
		}

		// Inject profile base if not provided
		if ra.BaseOverride == "" {
			cfg, _ := profile.Load()
			if cfg.Active != "" {
				if p, ok := cfg.Profiles[cfg.Active]; ok {
					ra.BaseOverride = p.Base
				}
			}
		}

		req, curl, err := openapi.BuildRequest(idx, spec, ra)
		if err != nil {
			fmt.Println("build error:", err)
			os.Exit(1)
		}

		if ra.ShowCurl && curl != "" {
			fmt.Println(curl)
		}

		resp, err := a.RunOnce(context.Background(), req)
		if err != nil {
			fmt.Println("request error:", err)
			os.Exit(1)
		}

		fmt.Printf("status: %d (dur=%dms)\n", resp.StatusCode, resp.DurationMs)
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
	fmt.Println("  restless-v2 openapi run <id> GET /pets --base https://petstore3.swagger.io/api/v3")
	fmt.Println("  restless-v2 openapi run <id> GET /pets/{petId} --base https://petstore3.swagger.io/api/v3 -p petId=7")
	fmt.Println("  restless-v2 openapi run <id> GET /pets --base https://petstore3.swagger.io/api/v3 -q limit=10 --curl")
	fmt.Println("  restless-v2 openapi run <id> GET /secure -H 'Authorization: Bearer {{token}}' -set token=abc --base https://example.com")
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
	// "Key: Value"
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
