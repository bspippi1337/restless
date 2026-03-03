package export

import (
	"encoding/json"
	"os"
)

type Request struct {
	Method string            `json:"method"`
	URL    string            `json:"url"`
	Header map[string]string `json:"header,omitempty"`
	Body   string            `json:"body,omitempty"`
}

func JSON(req Request, out string) error {
	b, _ := json.MarshalIndent(req, "", "  ")
	if out == "" {
		_, err := os.Stdout.Write(b)
		return err
	}
	return os.WriteFile(out, b, 0644)
}
