package loader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type LoadOptions struct {
	AllowRemoteRefs bool
}

func Load(ctx context.Context, ref string, opt LoadOptions) (*openapi3.T, error) {
	ldr := openapi3.NewLoader()
	ldr.IsExternalRefsAllowed = opt.AllowRemoteRefs

	ldr.ReadFromURIFunc = func(_ *openapi3.Loader, uri *openapi3.URI) ([]byte, error) {
		u := uri.String()
		switch {
		case strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://"):
			req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
			if err != nil {
				return nil, err
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode > 299 {
				return nil, fmt.Errorf("openapi fetch failed: %s", resp.Status)
			}
			return io.ReadAll(resp.Body)
		default:
			path := strings.TrimPrefix(u, "file://")
			if !filepath.IsAbs(path) {
				if abs, err := filepath.Abs(path); err == nil {
					path = abs
				}
			}
			return os.ReadFile(path)
		}
	}

	doc, err := ldr.LoadFromURI(refToURI(ref))
	if err != nil {
		return nil, err
	}
	if err := doc.Validate(ctx); err != nil {
		return nil, fmt.Errorf("openapi spec invalid: %w", err)
	}
	return doc, nil
}

func refToURI(ref string) *openapi3.URI {
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") || strings.HasPrefix(ref, "file://") {
		u, _ := openapi3.NewURI(ref)
		return u
	}
	abs, _ := filepath.Abs(ref)
	u, _ := openapi3.NewURI("file://" + abs)
	return u
}
