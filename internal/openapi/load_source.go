package openapi

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
)

func loadSource(src string) ([]byte, error) {

	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {

		client := &http.Client{}

		resp, err := client.Get(src)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		# detect HTML response
		if strings.HasPrefix(strings.TrimSpace(string(data)), "<") {
			return nil, errors.New("received HTML instead of OpenAPI spec")
		}

		return data, nil
	}

	return os.ReadFile(src)
}
