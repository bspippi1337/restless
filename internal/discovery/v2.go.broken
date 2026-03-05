package discovery

import (
<<<<<<< HEAD
	"context"
	"net/http"
	"strings"
	"time"
=======
	"github.com/bspippi1337/restless/internal/app"
"context"
"net/http"
"strings"
"time"
>>>>>>> parent of 0b706ce (Revert "engine-council: connect discovery, recon and probe engines to blackboard")

	"github.com/bspippi1337/restless/internal/httpx"
	"github.com/bspippi1337/restless/internal/store"
	"github.com/bspippi1337/restless/internal/util"
)

var common = []string{
	"/",
	"/api",
	"/v1",
	"/v2",
	"/users",
	"/repos",
	"/projects",
	"/orgs",
	"/teams",
	"/status",
	"/health",
	"/version",
}

func Discover(base string) (*store.API, error) {

	client := httpx.New()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	api := &store.API{
		BaseURL: base,
	}

	seen := map[string]bool{}

	for _, p := range common {

		url := util.JoinURL(base, p)

		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

		res, err := client.HTTP.Do(req)
		if err != nil {
			continue
		}

		res.Body.Close()

		if res.StatusCode < 400 {

			path := normalize(p)

			if !seen[path] {

				api.Endpoints = append(api.Endpoints, store.Endpoint{
					Path: path,
				})

				seen[path] = true
			}

		}

	}

	return api, nil
}

func normalize(p string) string {

	if p == "/" {
		return "/"
	}

<<<<<<< HEAD
	return strings.TrimRight(p, "/")
=======
url := util.JoinURL(base,p)

req,_ := http.NewRequestWithContext(ctx,"GET",url,nil)

res,err := client.HTTP.Do(req)
if err != nil {
continue
}

res.Body.Close()

if res.StatusCode < 400 {

path := normalize(p)
	app.PublishFinding("discovery","endpoint",path,"discovered path",0.7)

if !seen[path] {

api.Endpoints = append(api.Endpoints,store.Endpoint{
Path: path,
})

seen[path] = true
}

}

}

return api,nil
}

func normalize(p string) string{

if p == "/" {
return "/"
}

return strings.TrimRight(p,"/")
>>>>>>> parent of 0b706ce (Revert "engine-council: connect discovery, recon and probe engines to blackboard")
}
