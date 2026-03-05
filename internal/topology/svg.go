package topology

import (
	"html"
	"sort"
	"strings"
)

func SVG(host string, paths []string) string {
	uniq := map[string]bool{}
	clean := []string{}
	for _, p := range paths {
		p = "/" + strings.TrimLeft(strings.TrimSpace(p), "/")
		if p == "/" || p == "" {
			continue
		}
		if !uniq[p] {
			uniq[p] = true
			clean = append(clean, p)
		}
	}
	sort.Strings(clean)

	const (
		w      = 900
		lineH  = 18
		pad    = 18
		fontSz = 14
	)
	h := pad*2 + lineH*(len(clean)+2)

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="` + itoa(w) + `" height="` + itoa(h) + `" viewBox="0 0 ` + itoa(w) + ` ` + itoa(h) + `">` + "\n")
	b.WriteString(`<rect x="0" y="0" width="100%" height="100%" fill="#0b0f19"/>` + "\n")
	b.WriteString(`<text x="` + itoa(pad) + `" y="` + itoa(pad+lineH) + `" font-family="monospace" font-size="` + itoa(fontSz+2) + `" fill="#e6edf3">` + html.EscapeString(host) + `</text>` + "\n")

	y := pad + lineH*2
	for _, p := range clean {
		depth := strings.Count(strings.TrimLeft(p, "/"), "/")
		x := pad + depth*18
		label := "└ " + strings.TrimLeft(p, "/")
		b.WriteString(`<text x="` + itoa(x) + `" y="` + itoa(y) + `" font-family="monospace" font-size="` + itoa(fontSz) + `" fill="#a5b4fc">` + html.EscapeString(label) + `</text>` + "\n")
		y += lineH
	}

	b.WriteString(`</svg>` + "\n")
	return b.String()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [32]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + (n % 10))
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
