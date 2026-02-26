package tui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func Run(stdin, stdout *os.File, quiet bool) error {
	m := newModel(quiet)
	p := tea.NewProgram(
		m,
		tea.WithInput(stdin),
		tea.WithOutput(stdout),
		tea.WithAltScreen(),
	)
	_, err := p.Run()
	return err
}
