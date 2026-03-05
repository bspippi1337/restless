package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Request struct {
	w, h   int
	method textinput.Model
	url    textinput.Model
}

func NewRequest() Request {
	m := textinput.New()
	m.Prompt = "Method: "
	m.SetValue("GET")

	u := textinput.New()
	u.Prompt = "URL: "
	u.Placeholder = "https://api.example.com/health"

	return Request{method: m, url: u}
}

func (r *Request) SetSize(w, h int) { r.w, r.h = w, h }

func (r *Request) SetSuggestion(baseURL, method, path string) {
	if method != "" {
		r.method.SetValue(method)
	}
	if baseURL != "" && path != "" {
		r.url.SetValue(strings.TrimRight(baseURL, "/") + path)
	}
}

func (r Request) Update(msg tea.Msg) (Request, tea.Cmd) {
	var cmd tea.Cmd
	r.method, cmd = r.method.Update(msg)
	r.url, _ = r.url.Update(msg)
	return r, cmd
}

func (r Request) View() string {
	card := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).BorderForeground(cLine)
	title := lipgloss.NewStyle().Bold(true).Foreground(cTitle).Render("Request Builder")
	sub := lipgloss.NewStyle().Foreground(cDim).Render("v2 alpha: seeded by discovery. v2.1 adds send/response/presets.")

	body := strings.Join([]string{
		r.method.View(),
		r.url.View(),
		"",
		lipgloss.NewStyle().Foreground(cDim).Render("Next: add headers/body panel + send + response viewer."),
	}, "\n")

	return card.Render(title + "\n" + sub + "\n\n" + body)
}
