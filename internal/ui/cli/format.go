package cli

import (
	"fmt"
	"strings"
	"time"
)

func (s *State) PrintHeader(title string) {
	if s.NoHeader {
		return
	}
	fmt.Printf("Restless v0.2\n")
	if s.SessionName == "" {
		s.SessionName = "default"
	}
	fmt.Printf("Session: %s\n", s.SessionName)
	if s.Session.BaseURL != "" {
		fmt.Printf("Base:    %s\n", s.Session.BaseURL)
	}
	if s.Session.Mode != "" {
		fmt.Printf("Mode:    %s\n", s.Session.Mode)
	}
	fmt.Println(strings.Repeat("-", 40))
	if title != "" {
		fmt.Println(title)
	}
}

func formatDuration(d time.Duration) string {
	ms := float64(d.Microseconds()) / 1000.0
	if ms < 1000 {
		return fmt.Sprintf("%.0f ms", ms)
	}
	return fmt.Sprintf("%.2f s", ms/1000.0)
}
