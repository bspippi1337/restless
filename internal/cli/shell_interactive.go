package cli

import (
	"fmt"
	"strings"

	prompt "github.com/c-bata/go-prompt"
)

type InteractiveShell struct {
	api string
}

func NewInteractiveShell(api string) *InteractiveShell {
	return &InteractiveShell{api: api}
}

func (s *InteractiveShell) Run() {

	fmt.Println("restless interactive shell")
	fmt.Println("tab completion enabled")
	fmt.Println("type 'help' for commands")
	fmt.Println()

	p := prompt.New(
		s.executor,
		s.completer,
		prompt.OptionPrefix("restless> "),
	)

	p.Run()
}

func (s *InteractiveShell) executor(input string) {

	args := strings.Fields(input)

	if len(args) == 0 {
		return
	}

	switch args[0] {

	case "help":
		fmt.Println("commands:")
		fmt.Println(" discover <url>")
		fmt.Println(" map")
		fmt.Println(" inspect <endpoint>")
		fmt.Println(" call <METHOD> <url>")
		fmt.Println(" exit")

	case "exit", "quit":
		fmt.Println("bye")
		return

	case "map":
		fmt.Println("printing endpoint map")

	case "discover":
		if len(args) < 2 {
			fmt.Println("usage: discover <url>")
			return
		}
		fmt.Println("discovering api:", args[1])

	case "inspect":
		if len(args) < 2 {
			fmt.Println("usage: inspect <endpoint>")
			return
		}
		fmt.Println("inspect:", args[1])

	case "call":
		if len(args) < 3 {
			fmt.Println("usage: call GET /users")
			return
		}
		fmt.Println("calling", args[1], args[2])

	default:
		fmt.Println("unknown command:", args[0])
	}
}

func (s *InteractiveShell) completer(d prompt.Document) []prompt.Suggest {

	suggestions := []prompt.Suggest{
		{Text: "discover", Description: "scan API"},
		{Text: "map", Description: "print endpoint map"},
		{Text: "inspect", Description: "inspect endpoint"},
		{Text: "call", Description: "perform HTTP request"},
		{Text: "help", Description: "show help"},
		{Text: "exit", Description: "exit shell"},
	}

	return prompt.FilterHasPrefix(
		suggestions,
		d.GetWordBeforeCursor(),
		true,
	)
}
