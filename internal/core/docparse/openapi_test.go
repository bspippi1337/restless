package docparse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTryOpenAPIFindsPaths(t *testing.T) {
	spec := `openapi: "3.0.0"
paths:
  /health:
    get: {}
`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/openapi.yaml" {
			w.WriteHeader(200)
			w.Write([]byte(spec))
			return
		}
		w.WriteHeader(404)
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	o, _, err := TryOpenAPI(ctx, ts.URL)
	if err != nil || o == nil {
		t.Fatalf("expected openapi, got err=%v", err)
	}
	if len(o.Paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(o.Paths))
	}
}
