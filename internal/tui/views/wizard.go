package views

import (
	"fmt"
	"strings"

	"github.com/bspippi1337/restless/internal/core/discovery"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Wizard struct {
	w, h int

	domain textinput.Model
	last   *discovery.Finding
	err    string
}

func NewWizard() Wizard {
	d := textinput.New()
	d.Placeholder = "bankid.no / openai.com / api.example.com"
	d.Prompt = "Domain: "
	d.Focus()
	return Wizard{domain: d}
}

func (wz *Wizard) SetSize(w, h int) { wz.w, wz.h = w, h }

func (wz Wizard) DomainValue() string { return strings.TrimSpace(wz.domain.Value()) }

func (wz *Wizard) SetDiscovery(f *discovery.Finding, err string) {
	wz.last = f
	wz.err = err
}

func (wz Wizard) Update(msg tea.Msg) (Wizard, tea.Cmd) {
	var cmd tea.Cmd
	wz.domain, cmd = wz.domain.Update(msg)
	return wz, cmd
}

func (wz Wizard) View(busy bool) string {
	card := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).BorderForeground(cLine)
	title := lipgloss.NewStyle().Bold(true).Foreground(cTitle).Render("Connect & Discover (one input)")
	sub := lipgloss.NewStyle().Foreground(cDim).Render("Enter a domain. Press Ctrl+D to discover.")

	status := ""
	if busy {
		status = lipgloss.NewStyle().Foreground(cWarn).Render("Discovering… (docs → scrape → fuzz → verify)")
	} else if wz.err != "" {
		status = lipgloss.NewStyle().Foreground(cWarn).Render("Discovery error: " + wz.err)
	} else if wz.last != nil {
		status = lipgloss.NewStyle().Foreground(cOk).Render(fmt.Sprintf("Base: %s · Endpoints: %d", wz.last.BaseURL, len(wz.last.Endpoints)))
	}

	results := ""
	if wz.last != nil && len(wz.last.Endpoints) > 0 {
		lines := []string{lipgloss.NewStyle().Bold(true).Render("Top endpoints:")}
		maxn := 10
		if len(wz.last.Endpoints) < maxn {
			maxn = len(wz.last.Endpoints)
		}
		for i := 0; i < maxn; i++ {
			e := wz.last.Endpoints[i]
			lines = append(lines, fmt.Sprintf("  %s %s", e.Method, e.Path))
		}
		results = "\n\n" + lipgloss.NewStyle().Foreground(cDim).Render(strings.Join(lines, "\n"))
	}

	out := title + "\n" + sub + "\n\n" + wz.domain.View()
	if status != "" {
		out += "\n\n" + status
	}
	out += results
	out += "\n\n" + lipgloss.NewStyle().Foreground(cDim).Render("Tip: Press ? for help · Tab to switch views")
	return card.Render(out)
}
