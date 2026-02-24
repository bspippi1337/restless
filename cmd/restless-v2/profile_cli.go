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
