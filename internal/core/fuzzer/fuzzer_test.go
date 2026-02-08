package fuzzer

import (
	"testing"

	"github.com/bspippi1337/restless/internal/core/model"
)

func TestExpandIsBounded(t *testing.T) {
	seeds := []model.Endpoint{
		{Method: "GET", Path: "/v1/users"},
		{Method: "GET", Path: "/items/{id}"},
	}
	out := Expand(seeds, Options{MaxExtra: 10})
	if len(out) == 0 {
		t.Fatalf("expected some expansions")
	}
	if len(out) > 10 {
		t.Fatalf("expected bounded expansions, got %d", len(out))
	}
}
