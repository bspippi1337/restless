package openapi

import (
	"io"
	"net/http"
	"os"
	"strings"
)

func loadSource(src string) ([]byte, error) {

	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {

		resp, err := http.Get(src)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		return io.ReadAll(resp.Body)
	}

	return loadSource(src)
}
