package report

import (
	"encoding/json"
	"io"

	"github.com/bspippi1337/restless/internal/core"
)

type JSONOptions struct {
	Pretty bool
}

func WriteJSON(w io.Writer, result core.VerificationResult, opts JSONOptions) error {
	enc := json.NewEncoder(w)
	if opts.Pretty {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(result)
}
