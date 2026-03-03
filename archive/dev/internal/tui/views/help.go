package views

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Help struct {
	viewport viewport.Model
	ready    bool
	content  string
}

func NewHelp() Help {
	data, _ := os.ReadFile("docs/CLI-HELP.md")
	txt := string(data)
	if strings.TrimSpace(txt) == "" {
		txt = "Help file not found.\n\nExpected docs/CLI-HELP.md"
	}
	return Help{content: txt}
}

func (h *Help) SetSize(w, height int) {
	if !h.ready {
		h.viewport = viewport.New(w-4, height-6)
		h.viewport.SetContent(h.content)
		h.ready = true
		return
	}
	h.viewport.Width = w - 4
	h.viewport.Height = height - 6
}

func (h Help) Update(msg tea.Msg) (Help, tea.Cmd) {
	if !h.ready {
		return h, nil
	}
	var cmd tea.Cmd
	h.viewport, cmd = h.viewport.Update(msg)
	return h, cmd
}

func (h Help) View() string {
	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cLine).
		Padding(1, 2)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(cTitle).
		Render("Help · restless")

	return card.Render(title + "\n\n" + h.viewport.View())
}
