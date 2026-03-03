package loader

import (
	"context"
	"fmt"
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
	ldr.Context = ctx

	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		resp, err := http.Get(ref)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return nil, fmt.Errorf("openapi fetch failed: %s", resp.Status)
		}

		tmp, err := os.CreateTemp("", "restless-openapi-*.json")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tmp.Name())

		if _, err := tmp.ReadFrom(resp.Body); err != nil {
			return nil, err
		}
		tmp.Close()

		doc, err := ldr.LoadFromFile(tmp.Name())
		if err != nil {
			return nil, err
		}
		if err := doc.Validate(ctx); err != nil {
			return nil, err
		}
		return doc, nil
	}

	if !filepath.IsAbs(ref) {
		ref, _ = filepath.Abs(ref)
	}

	doc, err := ldr.LoadFromFile(ref)
	if err != nil {
		return nil, err
	}
	if err := doc.Validate(ctx); err != nil {
		return nil, err
	}

	return doc, nil
}
