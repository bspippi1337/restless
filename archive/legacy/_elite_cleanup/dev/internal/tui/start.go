package tui

import (
	"fmt"
	"io"
)

func Start(in io.Reader, out io.Writer) error {
	fmt.Fprintln(out, "🌀 Restless Interactive Mode")
	fmt.Fprintln(out, "Type 'help' or 'quit'")
	var input string
	for {
		fmt.Fprint(out, "> ")
		_, err := fmt.Fscanln(in, &input)
		if err != nil {
			return err
		}
		switch input {
		case "quit", "exit":
			fmt.Fprintln(out, "Bye 👋")
			return nil
		case "help":
			fmt.Fprintln(out, "Available commands: help, quit")
		default:
			fmt.Fprintf(out, "Unknown command: %s\n", input)
		}
	}
}
