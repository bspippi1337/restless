#!/usr/bin/env bash
set -euo pipefail

echo "==> Adding Profiles + OperationId runner"

mkdir -p internal/profile

# -------------------------
# PROFILE STORE
# -------------------------
cat > internal/profile/store.go <<'EOT'
package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Profile struct {
	Name string `json:"name"`
	Base string `json:"base"`
}

type Config struct {
	Active   string             `json:"active"`
	Profiles map[string]Profile `json:"profiles"`
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".restless", "profiles.json"), nil
}

func Load() (Config, error) {
	p, err := configPath()
	if err != nil {
		return Config{}, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return Config{Profiles: map[string]Profile{}}, nil
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return Config{}, err
	}
	if c.Profiles == nil {
		c.Profiles = map[string]Profile{}
	}
	return c, nil
}

func Save(c Config) error {
	p, err := configPath()
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(p), 0755)
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0644)
}
EOT

# -------------------------
# PROFILE CLI
# -------------------------
cat > cmd/restless-v2/profile_cli.go <<'EOT'
package main

import (
	"fmt"
	"os"

	"github.com/bspippi1337/restless/internal/profile"
)

func handleProfile(args []string) {
	if len(args) < 1 {
		fmt.Println("profile <set|use|ls>")
		os.Exit(1)
	}

	cfg, _ := profile.Load()

	switch args[0] {

	case "set":
		if len(args) < 3 {
			fmt.Println("profile set <name> base=<url>")
			os.Exit(1)
		}
		name := args[1]
		base := ""
		for _, a := range args[2:] {
			if len(a) > 5 && a[:5] == "base=" {
				base = a[5:]
			}
		}
		if base == "" {
			fmt.Println("missing base=")
			os.Exit(1)
		}
		cfg.Profiles[name] = profile.Profile{
			Name: name,
			Base: base,
		}
		profile.Save(cfg)
		fmt.Println("saved profile:", name)

	case "use":
		if len(args) < 2 {
			fmt.Println("profile use <name>")
			os.Exit(1)
		}
		name := args[1]
		if _, ok := cfg.Profiles[name]; !ok {
			fmt.Println("profile not found")
			os.Exit(1)
		}
		cfg.Active = name
		profile.Save(cfg)
		fmt.Println("active profile:", name)

	case "ls":
		for k, v := range cfg.Profiles {
			active := ""
			if k == cfg.Active {
				active = "*"
			}
			fmt.Printf("%s %s base=%s\n", active, k, v.Base)
		}

	default:
		fmt.Println("unknown profile command")
	}
}
EOT

# -------------------------
# PATCH MAIN DISPATCH
# -------------------------
sed -i '/case "openapi":/a\
        case "profile":\
            handleProfile(os.Args[2:])\
            return' cmd/restless-v2/main.go

# -------------------------
# OPERATION ID SUPPORT
# -------------------------
cat >> internal/modules/openapi/run.go <<'EOT'

// Resolve operationId to method/path
func ResolveOperationID(spec Spec, opID string) (string, string, error) {
	for path, item := range spec.Paths {
		for method, op := range item {
			if op.OperationID == opID {
				return strings.ToUpper(method), path, nil
			}
		}
	}
	return "", "", errors.New("operationId not found")
}
EOT

# Patch run handler for opID fallback
sed -i '/idx, err := openapi.LoadIndex/a\
        // OperationId fallback\
        if ra.Path == "" {\
            m, p, err := openapi.ResolveOperationID(spec, ra.Method)\
            if err == nil {\
                ra.Method = m\
                ra.Path = p\
            }\
        }' cmd/restless-v2/openapi_cli.go || true

gofmt -w .
go test ./...

echo "Profiles + OperationId installed."
