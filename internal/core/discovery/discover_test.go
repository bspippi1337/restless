package discovery

import "testing"

func TestDiscoverDomain_Empty(t *testing.T) {
	if _, err := DiscoverDomain("", Options{}); err == nil {
		t.Fatalf("expected error for empty domain")
	}
}

func TestDiscoverDomain_Basic(t *testing.T) {
	find, err := DiscoverDomain("openai.com", Options{Verify: false, Fuzz: true, BudgetSeconds: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if find.Domain != "openai.com" {
		t.Fatalf("domain mismatch: got %q", find.Domain)
	}
	if len(find.BaseURLs) == 0 || len(find.DocURLs) == 0 || len(find.Endpoints) == 0 {
		t.Fatalf("expected baseUrls/docUrls/endpoints to be non-empty: %#v", find)
	}
	if find.Endpoints[0].Method == "" || find.Endpoints[0].Path == "" {
		t.Fatalf("expected endpoint method+path set: %#v", find.Endpoints[0])
	}
}

func TestDiscoverDomain_VerifyDoesNotCrash(t *testing.T) {
	// We don't assert network behavior (CI may run offline).
	find, err := DiscoverDomain("example.invalid", Options{Verify: true, BudgetSeconds: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if find.Domain != "example.invalid" {
		t.Fatalf("domain mismatch: %q", find.Domain)
	}
}
