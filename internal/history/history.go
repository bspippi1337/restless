package history

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/bspippi1337/restless/internal/core"
)

const dir = ".restless"
const lastFile = "last-run.json"

func Save(result core.VerificationResult) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, lastFile)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func Load() (*core.VerificationResult, error) {
	path := filepath.Join(dir, lastFile)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var r core.VerificationResult
	if err := json.NewDecoder(f).Decode(&r); err != nil {
		return nil, err
	}

	return &r, nil
}
