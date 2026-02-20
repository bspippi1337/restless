package console

import (
	"strings"
	"testing"

	"github.com/bspippi1337/restless/internal/httpclient"
)

func TestSplitArgs(t *testing.T) {
	got := splitArgs(`save "My Snip" --flag`)
	want := []string{"save", "My Snip", "--flag"}
	if len(got) != len(want) {
		t.Fatalf("len mismatch: %#v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d mismatch: got=%q want=%q", i, got[i], want[i])
		}
	}
}

func TestToCurl_EscapesQuotes(t *testing.T) {
	r := httpclient.Request{
		Method: "POST",
		Headers: map[string]string{
			"Authorization": "Bearer token",
		},
		Body: []byte(`{"msg":"it's fine"}`),
	}
	out := toCurl("https://example.com/v1", r)
	if !strings.Contains(out, "curl -i") || !strings.Contains(out, "-X POST") {
		t.Fatalf("unexpected curl: %s", out)
	}
	// Expected POSIX-safe escape for single quotes inside a single-quoted string.
	if !strings.Contains(out, `'"'"'`) {
		t.Fatalf("expected POSIX quote-escape sequence in output, got: %s", out)
	}
}
