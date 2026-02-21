package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bspippi1337/restless/internal/core/discovery"
	"github.com/bspippi1337/restless/internal/tui/views"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type tabs int

const (
	tabWizard tabs = iota
	tabRequest
	tabStream
	tabHelp
)

type keymap struct {
	TabNext  key.Binding
	TabPrev  key.Binding
	Quit     key.Binding
	Discover key.Binding
	Help     key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.Discover, k.Help, k.TabPrev, k.TabNext, k.Quit}
}

func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Discover, k.Help},
		{k.TabPrev, k.TabNext},
		{k.Quit},
	}
}

type model struct {
	quiet bool
	w, h  int
	tab   tabs

	help help.Model
	keys keymap

	wizard views.Wizard
	req    views.Request
	stream views.Stream
	helpv  views.Help
	face   views.Face

	discoverBusy bool
	discoverErr  string
}

type tickMsg time.Time

type discoverMsg struct {
	finding discovery.Finding
	err     error
}

func newModel(quiet bool) model {
	k := keymap{
		TabNext:  key.NewBinding(key.WithKeys("tab", "l"), key.WithHelp("tab", "next tab")),
		TabPrev:  key.NewBinding(key.WithKeys("shift+tab", "h"), key.WithHelp("shift+tab", "prev tab")),
		Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Discover: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "discover")),
		Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	}

	return model{
		quiet:  quiet,
		tab:    tabWizard,
		help:   help.New(),
		keys:   k,
		wizard: views.NewWizard(),
		req:    views.NewRequest(),
		stream: views.NewStream(),
		helpv:  views.NewHelp(),
		face:   views.NewFace(quiet),
	}
}

func (m model) Init() tea.Cmd {
	if m.quiet {
		return nil
	}
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		hh := max(10, m.h-6)
		m.wizard.SetSize(m.w, hh)
		m.req.SetSize(m.w, hh)
		m.stream.SetSize(m.w, hh)
		m.helpv.SetSize(m.w, hh)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.tab = tabHelp
			return m, nil
		case key.Matches(msg, m.keys.TabNext):
			m.tab = (m.tab + 1) % 4
			return m, nil
		case key.Matches(msg, m.keys.TabPrev):
			m.tab = (m.tab + 3) % 4
			return m, nil
		case key.Matches(msg, m.keys.Discover):
			if m.tab == tabWizard && !m.discoverBusy {
				m.discoverBusy = true
				m.discoverErr = ""
				domain := m.wizard.DomainValue()
				return m, m.runDiscover(domain)
			}
		}

	case discoverMsg:
		m.discoverBusy = false
		if msg.err != nil {
			m.discoverErr = msg.err.Error()
			m.wizard.SetDiscovery(nil, m.discoverErr)
			return m, nil
		}

		m.wizard.SetDiscovery(&msg.finding, "")
		if len(msg.finding.Endpoints) > 0 {
			m.req.SetSuggestion(
				msg.finding.BaseURL,
				msg.finding.Endpoints[0].Method,
				msg.finding.Endpoints[0].Path,
			)
		}
		return m, nil

	case tickMsg:
		if !m.quiet {
			m.face.Tick()
			return m, tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
		}
	}

	switch m.tab {
	case tabWizard:
		var cmd tea.Cmd
		m.wizard, cmd = m.wizard.Update(msg)
		return m, cmd
	case tabRequest:
		var cmd tea.Cmd
		m.req, cmd = m.req.Update(msg)
		return m, cmd
	case tabStream:
		var cmd tea.Cmd
		m.stream, cmd = m.stream.Update(msg)
		return m, cmd
	case tabHelp:
		var cmd tea.Cmd
		m.helpv, cmd = m.helpv.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m model) runDiscover(domain string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 18*time.Second)
		defer cancel()

		_ = ctx // beholdt, i tilfelle discovery tar ctx senere

		finding, err := discovery.DiscoverDomain(domain, discovery.Options{
			BudgetSeconds: 15,
			BudgetPages:   6,
			Verify:        true,
			Fuzz:          true,
		})
		return discoverMsg{finding: finding, err: err}
	}
}

func (m model) View() string {
	header := views.Header("restless", m.tabLabel(), m.face.View())
	body := m.activeView()
	footer := views.Footer(m.help.View(m.keys))
	return strings.Join([]string{header, body, footer}, "\n")
}

func (m model) tabLabel() string {
	switch m.tab {
	case tabWizard:
		return "Connect & Discover"
	case tabRequest:
		return "Request Builder"
	case tabStream:
		return "Stream (SSE)"
	case tabHelp:
		return "Help"
	default:
		return fmt.Sprintf("Tab %d", m.tab)
	}
}

func (m model) activeView() string {
	switch m.tab {
	case tabWizard:
		return m.wizard.View(m.discoverBusy)
	case tabRequest:
		return m.req.View()
	case tabStream:
		return m.stream.View()
	case tabHelp:
		return m.helpv.View()
	default:
		return ""
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
