package discovery

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscoverDomainAgainstLocalServer(t *testing.T) {
	spec := `openapi: "3.0.0"
paths:
  /health:
    get: {}
`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/openapi.yaml":
			w.WriteHeader(200)
			w.Write([]byte(spec))
		case "/health":
			w.WriteHeader(200)
		default:
			w.WriteHeader(404)
		}
	}))
	defer ts.Close()

	find, err := DiscoverDomain(ts.URL, Options{BudgetSeconds: 5, BudgetPages: 2, Verify: true, Fuzz: true})
	if err != nil {
		t.Fatalf("discover error: %v", err)
	}
	if len(find.Endpoints) == 0 {
		t.Fatalf("expected endpoints")
	}
}
