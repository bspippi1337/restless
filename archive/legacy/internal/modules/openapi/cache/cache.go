package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/bspippi1337/restless/internal/modules/openapi/auto"
	gloader "github.com/bspippi1337/restless/internal/modules/openapi/guard/loader"
)

type cachedSpec struct {
	Doc      *openapi3.T
	SpecRef  string
	LoadedAt time.Time
}

var (
	mu    sync.RWMutex
	store = map[string]*cachedSpec{}
	ttl   = 10 * time.Minute
)

func cacheDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".cache", "restless", "openapi")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func cachePath(baseURL string) string {
	h := sha256.Sum256([]byte(baseURL))
	return filepath.Join(cacheDir(), hex.EncodeToString(h[:])+".json")
}

// Get returns an OpenAPI doc for baseURL using memory cache, then disk cache, then network refresh.
// If offline/unreachable, disk cache is used as fallback.
func Get(ctx context.Context, baseURL string) (*openapi3.T, string, bool) {
	// 1) memory
	mu.RLock()
	entry, exists := store[baseURL]
	mu.RUnlock()

	if exists && time.Since(entry.LoadedAt) < ttl {
		return entry.Doc, entry.SpecRef, true
	}

	// 2) network refresh
	if specRef, ok := auto.TryDiscover(baseURL); ok {
		doc, err := gloader.Load(ctx, specRef, gloader.LoadOptions{AllowRemoteRefs: true})
		if err == nil {
			saveToDisk(baseURL, doc, specRef)
			mu.Lock()
			store[baseURL] = &cachedSpec{Doc: doc, SpecRef: specRef, LoadedAt: time.Now()}
			mu.Unlock()
			return doc, specRef, true
		}
	}

	// 3) disk fallback
	doc, specRef, ok := loadFromDisk(baseURL)
	if ok {
		mu.Lock()
		store[baseURL] = &cachedSpec{Doc: doc, SpecRef: specRef, LoadedAt: time.Now()}
		mu.Unlock()
		return doc, specRef, true
	}

	return nil, "", false
}

type diskEntry struct {
	SpecRef string          `json:"spec_ref"`
	Raw     json.RawMessage `json:"raw_spec"`
}

func saveToDisk(baseURL string, doc *openapi3.T, specRef string) {
	raw, err := json.Marshal(doc)
	if err != nil {
		return
	}
	entry := diskEntry{SpecRef: specRef, Raw: raw}
	b, err := json.Marshal(entry)
	if err != nil {
		return
	}
	_ = os.WriteFile(cachePath(baseURL), b, 0644)
}

func loadFromDisk(baseURL string) (*openapi3.T, string, bool) {
	b, err := os.ReadFile(cachePath(baseURL))
	if err != nil {
		return nil, "", false
	}
	var entry diskEntry
	if err := json.Unmarshal(b, &entry); err != nil {
		return nil, "", false
	}
	var doc openapi3.T
	if err := json.Unmarshal(entry.Raw, &doc); err != nil {
		return nil, "", false
	}
	return &doc, entry.SpecRef, true
}

func Invalidate(baseURL string) {
	mu.Lock()
	delete(store, baseURL)
	mu.Unlock()
	_ = os.Remove(cachePath(baseURL))
}
