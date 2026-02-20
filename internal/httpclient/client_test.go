package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestBuildURL(t *testing.T) {
	u, err := BuildURL("https://example.com/api", "v1/status", map[string]string{"a": "1", "b": "2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(u, "https://example.com/api/v1/status?") {
		t.Fatalf("unexpected url: %s", u)
	}
	if !strings.Contains(u, "a=1") || !strings.Contains(u, "b=2") {
		t.Fatalf("missing query params: %s", u)
	}
}

func TestDo_Smoke(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(405)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"echo":"` + r.URL.Path + `"}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := Do(ctx, Request{
		Method:  "POST",
		BaseURL: srv.URL,
		Path:    "/hello",
		Headers: map[string]string{"X-Test": "1"},
		Body:    []byte(`{"hi":"there"}`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("unexpected status: %d", res.StatusCode)
	}
	if !strings.Contains(string(res.Body), `"ok":true`) {
		t.Fatalf("unexpected body: %s", string(res.Body))
	}
	if res.LatencyMs < 0 {
		t.Fatalf("latency not set")
	}
}

func TestRedact(t *testing.T) {
	old := os.Getenv("RESTLESS_TOKEN")
	old2 := os.Getenv("OPENAI_API_KEY")
	t.Cleanup(func() {
		_ = os.Setenv("RESTLESS_TOKEN", old)
		_ = os.Setenv("OPENAI_API_KEY", old2)
	})

	_ = os.Setenv("RESTLESS_TOKEN", "sk-THIS_IS_A_LONG_TOKEN")
	in := "Authorization: Bearer sk-THIS_IS_A_LONG_TOKEN"
	out := Redact(in)
	if out == in {
		t.Fatalf("expected redaction")
	}
	if strings.Contains(out, "THIS_IS_A_LONG_TOKEN") {
		t.Fatalf("expected token to be hidden: %s", out)
	}
}

func TestPrettyJSON(t *testing.T) {
	b := PrettyJSON([]byte(`{"a":1,"b":{"c":2}}`))
	if !strings.Contains(string(b), "\n") {
		t.Fatalf("expected indented json, got: %s", string(b))
	}
}
