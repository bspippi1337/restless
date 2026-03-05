package cli

import (
	"fmt"

	"github.com/bspippi1337/restless/internal/discovery"
	"github.com/bspippi1337/restless/internal/store"
	"github.com/spf13/cobra"
)

func NewLearnCmd() *cobra.Command {

	return &cobra.Command{
		Use:   "learn <url>",
		Short: "Discover API and store endpoints",
		Args:  cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {

			base := args[0]

			api, err := discovery.Discover(base)
			if err != nil {
				return err
			}

			cacheRoot, _ := cmd.Root().PersistentFlags().GetString("cache")
			cacheRoot, _ = store.DefaultRoot(cacheRoot)

			_, err = store.Write(cacheRoot, api)
			if err != nil {
				return err
			}

			fmt.Println("API learned:", base)
			fmt.Println("Endpoints discovered:", len(api.Endpoints))

			return nil
		},
	}
}
