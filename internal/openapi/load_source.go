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

		req, err := http.NewRequest("GET", src, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Accept", "application/json, application/yaml, */*")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		body := strings.TrimSpace(string(data))
		// hvis HTML, prøv fallback uten petstore redirect
		if strings.HasPrefix(body, "<") {
			return nil, errors.New("received HTML instead of OpenAPI spec")
		}

		return data, nil
	}

	return os.ReadFile(src)
}
