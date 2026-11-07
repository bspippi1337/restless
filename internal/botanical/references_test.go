package botanical

import "testing"

func TestReferenceSearchIsProvenanceBound(t *testing.T) {
	c := Catalog{Documents: []Document{
		{URL: "https://example.org/papers/auth.pdf", SHA256: "abc", Title: "Authorization whitepaper", Kind: "pdf-unparsed"},
		{URL: "https://example.org/docs/api.md", SHA256: "def", Title: "API reference", Kind: "reference", Terms: []string{"authorization", "session", "endpoint"}},
	}}
	hits, err := RankReferences(c, "authorization", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 {
		t.Fatalf("got %d results", len(hits))
	}
	if hits[0].Document.URL != "https://example.org/papers/auth.pdf" {
		t.Fatalf("unexpected ranking %+v", hits)
	}
	if hits[0].Document.SHA256 != "abc" {
		t.Fatal("lost evidence hash")
	}
	if _, err := RankReferences(c, "    ", 2); err == nil {
		t.Fatal("accepted empty query")
	}
	if _, err := RankReferences(c, "authorization", 51); err == nil {
		t.Fatal("accepted excessive result limit")
	}
}
