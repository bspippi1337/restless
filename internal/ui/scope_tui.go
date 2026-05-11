//go:build tui
// +build tui

package ui

import (
	"fmt"
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Signal struct {
	OK        int
	Redirect  int
	ClientErr int
	ServerErr int
	RateLimit int
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

type model struct {
	w, h  int
	phase float64
	sig   Signal
	title lipgloss.Style
	neon  lipgloss.Style
	muted lipgloss.Style
}

func NewScopeModel(sig Signal) model {
	return model{
		w: 100, h: 22, phase: 0, sig: sig,
		title: lipgloss.NewStyle().Foreground(lipgloss.Color("#00E0FF")).Bold(true),
		neon:  lipgloss.NewStyle().Foreground(lipgloss.Color("#00E0FF")),
		muted: lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7A89")),
	}
}

func (m model) Init() tea.Cmd { return tick() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tickMsg:
		m.phase += 0.15
		return m, tick()
	case tea.KeyMsg:
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.w = msg.Width - 4
		m.h = msg.Height - 6
		if m.h < 8 {
			m.h = 8
		}
		return m, nil
	}
	return m, nil
}

// Braille plotting (2x4 subpixels)
var dotMap = map[[2]int]int{
	{0, 0}: 1, {0, 1}: 2, {0, 2}: 3,
	{1, 0}: 4, {1, 1}: 5, {1, 2}: 6,
	{0, 3}: 7, {1, 3}: 8,
}

func braille(grid [][]bool, w, h int) []string {
	out := []string{}
	for y := 0; y < h; y += 4 {
		line := ""
		for x := 0; x < w; x += 2 {
			bits := 0
			for dy := 0; dy < 4; dy++ {
				for dx := 0; dx < 2; dx++ {
					px, py := x+dx, y+dy
					if py < h && px < w && grid[py][px] {
						bits |= 1 << (dotMap[[2]int{dx, dy}] - 1)
					}
				}
			}
			if bits == 0 {
				line += " "
			} else {
				line += string(rune(0x2800 + bits))
			}
		}
		out = append(out, line)
	}
	return out
}

func (m model) View() string {
	w, h := m.w, m.h
	if w < 20 {
		w = 20
	}
	if h < 8 {
		h = 8
	}

	grid := make([][]bool, h)
	for i := range grid {
		grid[i] = make([]bool, w)
	}
	// signal parameters derived from API stats
	base := 0.18
	jitter := float64(m.sig.ClientErr) * 0.02
	dist := float64(m.sig.ServerErr) * 0.06
	spike := float64(m.sig.RateLimit) * 0.22
	drift := float64(m.sig.Redirect) * 0.01

	for x := 0; x < w; x++ {
		t := float64(x)*base + m.phase
		y := math.Sin(t + jitter*math.Sin(t*1.7))
		y += dist * math.Sin(t*0.45)
		y += spike * math.Sin(t*2.8)
		y += drift * math.Cos(t*0.9)

		pos := int((y + 1) * float64(h-1) / 2)
		if pos < 0 {
			pos = 0
		}
		if pos >= h {
			pos = h - 1
		}

		grid[pos][x] = true
		if x%3 == 0 && pos+1 < h {
			grid[pos+1][x] = true
		}
	}

	lines := braille(grid, w, h)

	header := m.title.Render("⚡ restless signal monitor")
	stats := m.muted.Render(
		fmt.Sprintf("2xx:%d  3xx:%d  4xx:%d  5xx:%d  429:%d  (press any key to quit)",
			m.sig.OK, m.sig.Redirect, m.sig.ClientErr, m.sig.ServerErr, m.sig.RateLimit,
		),
	)

	body := ""
	for _, ln := range lines {
		body += m.neon.Render(ln) + "\n"
	}

	return header + "\n" + stats + "\n\n" + body
}

func RunScope(sig Signal) error {
	p := tea.NewProgram(NewScopeModel(sig), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
