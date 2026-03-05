package views

import "github.com/charmbracelet/lipgloss"

var (
	cTitle = lipgloss.Color("63")
	cDim   = lipgloss.Color("244")
	cLine  = lipgloss.Color("238")
	cOk    = lipgloss.Color("42")
	cWarn  = lipgloss.Color("214")
)

func Header(app, tab, face string) string {
	left := lipgloss.NewStyle().Bold(true).Foreground(cTitle).Render(app)
	mid := lipgloss.NewStyle().Foreground(cDim).Render(" · " + tab)
	right := lipgloss.NewStyle().Foreground(cDim).Render(face)

	return lipgloss.NewStyle().Padding(0, 1).Render(left+mid) + "  " + right + "\n" +
		lipgloss.NewStyle().Foreground(cLine).Render(lipgloss.NewStyle().Padding(0, 1).Render(repeat("─", 80)))
}

func Footer(help string) string {
	return lipgloss.NewStyle().Padding(0, 1).Foreground(cDim).Render(help)
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
