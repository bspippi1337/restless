package fuzz

import (
	"net/http"
	"time"
)

func Run(base string) ([]string, error) {

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	wordlist := []string{
		"/admin",
		"/login",
		"/users",
		"/metrics",
		"/debug",
		"/internal",
	}

	var found []string

	for _, w := range wordlist {

		resp, err := client.Get(base + w)
		if err == nil && resp != nil {

			resp.Body.Close()

			if resp.StatusCode != 404 {
				found = append(found, w)
			}
		}
	}

	return found, nil
}
