package cli

import (
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bspippi1337/restless/internal/httpx"
	"github.com/bspippi1337/restless/internal/store"
	"github.com/bspippi1337/restless/internal/util"
)

func AddDynamicCommands(root *cobra.Command) error {

	cacheRoot, _ := root.PersistentFlags().GetString("cache")
	apiName, _ := root.PersistentFlags().GetString("api")

	cacheRoot, _ = store.DefaultRoot(cacheRoot)

	api, err := store.Read(cacheRoot, apiName)
	if err != nil || api == nil {
		return nil
	}

	client := httpx.New()

	for _, e := range api.Endpoints {

		name := strings.TrimPrefix(e.Path, "/")
		name = strings.ReplaceAll(name, "/", "_")

		if name == "" {
			continue
		}

		path := e.Path

		cmd := &cobra.Command{
			Use:   name,
			Short: "GET " + path,

			RunE: func(cmd *cobra.Command, args []string) error {

				url := util.JoinURL(api.BaseURL, path)

				req, _ := http.NewRequest("GET", url, nil)

				res, err := client.HTTP.Do(req)
				if err != nil {
					return err
				}

				body, _ := httpx.ReadBody(res, 10<<20)

				cmd.Println(string(body))

				return nil
			},
		}

		root.AddCommand(cmd)

	}

	return nil
}
