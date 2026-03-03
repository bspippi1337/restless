package discovery

import "testing"

func TestHostCandidates(t *testing.T) {
	h := HostCandidates("example.com")
	if len(h) < 3 {
		t.Fatalf("expected candidates, got %d", len(h))
	}
	if h[0] != "https://example.com" {
		t.Fatalf("unexpected first: %s", h[0])
	}
}
