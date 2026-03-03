package discovery

import "net/http"

var candidates = []string{
	"/swagger.json",
	"/openapi.json",
	"/v3/api-docs",
	"/api-docs",
	"/swagger/v1/swagger.json",
}

func Find(base string) (string, bool) {

	for _, p := range candidates {

		u := base + p

		resp, err := http.Get(u)
		if err != nil {
			continue
		}

		resp.Body.Close()

		if resp.StatusCode == 200 {
			return u, true
		}
	}

	return "", false
}
