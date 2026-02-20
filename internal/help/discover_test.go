package help

import (
	"strings"
	"testing"
)

func TestDiscoverHelp_Output(t *testing.T) {
	h := DiscoverHelp(HelpContext{
		TerminalWidth: 80,
		LastDomain:    "openai.com",
		ActiveProfile: "openai",
		ProfileDir:    "/tmp/profiles",
		Profiles:      []string{"openai", "demo"},
		SupportsJSON:  true,
		SupportsTUI:   true,
	})
	mustContain := []string{
		"restless discover",
		"Usage:",
		"Examples",
		"--verify",
		"--fuzz",
		"--save-profile",
		"--json",
	}
	for _, s := range mustContain {
		if !strings.Contains(h, s) {
			t.Fatalf("help missing %q\n---\n%s\n---", s, h)
		}
	}
}
