package openapi

import (
	"io"
	"net/http"
	"os"
	"strings"
)

func LoadSource(path string) ([]byte, error) {

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {

		resp, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		return io.ReadAll(resp.Body)
	}

	return os.ReadFile(path)
}
