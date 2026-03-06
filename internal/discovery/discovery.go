package discovery

import (
	"encoding/json"
	"io"
	"net/http"
	neturl "net/url"
	"strings"
	"sync"

	"github.com/bspippi1337/restless/internal/store"
	"github.com/bspippi1337/restless/internal/ui"
	"github.com/bspippi1337/restless/internal/util"
)

func Discover(base string) []store.Endpoint {

	client := &http.Client{}

	queue := []string{"/"}

	seen := map[string]bool{}

	var endpoints []store.Endpoint

	for len(queue) > 0 {

		path := queue[0]
		queue = queue[1:]

		if seen[path] {
			continue
		}

		seen[path] = true

		url := util.JoinURL(base, path)

		ui.IncRequest()

		res, err := client.Get(url)

		if err != nil {
			continue
		}

		body, _ := io.ReadAll(res.Body)
		res.Body.Close()

		if res.StatusCode < 400 {

			ui.IncEndpoint()

			endpoints = append(endpoints, store.Endpoint{
				Path: path,
			})

		}

		var data any

		if json.Unmarshal(body, &data) != nil {
			continue
		}

		var walk func(any)

		walk = func(v any) {

			switch t := v.(type) {

			case map[string]any:

				for _, x := range t {
					walk(x)
				}

			case []any:

				for _, x := range t {
					walk(x)
				}

			case string:

				s := strings.TrimSpace(t)

				if strings.HasPrefix(s, "http") {

					u, err := neturl.Parse(s)

					if err == nil {

						if strings.Contains(base, u.Host) {

							queue = append(queue, u.Path)

						}

					}

				}

				if strings.HasPrefix(s, "/") {
					queue = append(queue, s)
				}

			}

		}

		walk(data)

	}

	return endpoints
}
