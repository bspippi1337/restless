package cli

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "restless",
		Short: "API discovery and probing tool",
	}

	cmd.AddCommand(NewAutoCmd())
	cmd.AddCommand(NewSmartCmd())
	cmd.AddCommand(NewSwarmCmd())

	cmd.AddCommand(NewDiscoverCmd())
	cmd.AddCommand(NewMapCmd())

	cmd.AddCommand(NewDiscoverCmd())
	cmd.AddCommand(NewMapCmd())

	cmd.AddCommand(NewDiscoverCmd())
	cmd.AddCommand(NewMapCmd())

	cmd.AddCommand(NewDiscoverCmd())
	cmd.AddCommand(NewMapCmd())

	return cmd
}
