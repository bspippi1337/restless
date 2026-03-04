package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {

	file := "internal/cli/root.go"

	data, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}

	src := string(data)

	add := []string{
		"cmd.AddCommand(NewAutoCmd())",
		"cmd.AddCommand(NewSmartCmd())",
		"cmd.AddCommand(NewSwarmCmd())",
	}

	insert := ""

	for _, c := range add {
		if !strings.Contains(src, c) {
			insert += "\n\t" + c
		}
	}

	if insert == "" {
		fmt.Println("CLI commands already registered")
		return
	}

	i := strings.Index(src, "cmd := &cobra.Command")
	if i == -1 {
		panic("could not find cobra root command")
	}

	lines := strings.Split(src, "\n")

	out := []string{}
	inserted := false

	for _, l := range lines {

		out = append(out, l)

		if !inserted && strings.Contains(l, "cmd := &cobra.Command") {
			inserted = true
			out = append(out, insert)
		}
	}

	err = os.WriteFile(file, []byte(strings.Join(out, "\n")), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("CLI commands registered successfully")
}
