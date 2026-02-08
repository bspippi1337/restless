package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Stream struct{ w, h int }

func NewStream() Stream { return Stream{} }
func (s *Stream) SetSize(w, h int) { s.w, s.h = w, h }
func (s Stream) Update(msg tea.Msg) (Stream, tea.Cmd) { return s, nil }

func (s Stream) View() string {
	card := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).BorderForeground(cLine)
	title := lipgloss.NewStyle().Bold(true).Foreground(cTitle).Render("Stream (SSE)")
	sub := lipgloss.NewStyle().Foreground(cDim).Render("v2 alpha: UI ready. v2.1 wires SSE engine + transcript export.")
	return card.Render(title + "\n" + sub + "\n\n" +
		lipgloss.NewStyle().Foreground(cDim).Render("Soon: connect SSE endpoint, show events live, reconnect on drop."))
}
