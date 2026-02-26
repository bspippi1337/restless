package cli

import (
	"context"
	"fmt"
	"github.com/bspippi1337/restless/internal/adapters/http"
	"github.com/bspippi1337/restless/internal/app"
	"github.com/bspippi1337/restless/internal/core/engine"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "restless",
	}

	cmd.AddCommand(newRunCmd())
	return cmd
}

func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run [method] [url]",
		Short: "Run a request using v2 engine",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			transport := &httpadapter.HTTPTransport{}
			eng := &engine.Engine{Transport: transport}
			runner := &app.Runner{Engine: eng}

			res := runner.Run(context.Background(), args[0], args[1])
			if res.Err != nil {
				fmt.Println("Error:", res.Err)
				return
			}

			fmt.Println("Status:", res.Status)
			fmt.Println(string(res.Body))
		},
	}
}
