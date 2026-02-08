package probe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestVerifyTreatsAuthAsOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ok, _, hint := Verify(ctx, "GET", ts.URL+"/secure")
	if !ok {
		t.Fatalf("expected ok, got false")
	}
	if hint != "auth required" {
		t.Fatalf("expected auth required hint, got %q", hint)
	}
}

func TestVerifySkips500ButAccepts200(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ok, _, _ := Verify(ctx, "GET", ts.URL+"/ok")
	if !ok {
		t.Fatalf("expected ok")
	}
}
