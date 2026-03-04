//go:build !tui
// +build !tui

package ui

import "fmt"

func RunScopeUI() error {
	fmt.Println("TUI disabled (build with -tags tui to enable)")
	return nil
}
