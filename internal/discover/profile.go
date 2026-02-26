package discover

import (
	"encoding/json"
	"net/http"
	"time"
)

type Profile struct {
	URL          string   `json:"url"`
	Methods      []string `json:"methods,omitempty"`
	ContentTypes []string `json:"content_types,omitempty"`
	DiscoveredAt string   `json:"discovered_at"`
}

func Probe(url string) (Profile, error) {
	p := Profile{
		URL:          url,
		DiscoveredAt: time.Now().UTC().Format(time.RFC3339),
	}

	resp, err := http.Head(url)
	if err != nil {
		return p, err
	}
	defer resp.Body.Close()

	if allow := resp.Header.Get("Allow"); allow != "" {
		p.Methods = []string{allow}
	}
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		p.ContentTypes = []string{ct}
	}

	return p, nil
}

func (p Profile) JSON() []byte {
	b, _ := json.MarshalIndent(p, "", "  ")
	return b
}
