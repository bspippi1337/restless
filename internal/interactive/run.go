package interactive

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bspippi1337/restless/internal/app"
)

func Run() int {
	fmt.Println("🌀 Restless Interactive Mode")
	fmt.Println("Type 'help' or 'quit'")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			return 0
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if line == "quit" || line == "exit" {
			return 0
		}

		args := strings.Fields(line)
		code := app.Run(args, os.Stdin, os.Stdout, os.Stderr)

		if code != 0 {
			fmt.Println("Command failed with exit code:", code)
		}
	}
}
