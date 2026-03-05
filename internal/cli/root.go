package cli

import "github.com/spf13/cobra"

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restless",
		Short: "API discovery and probing tool",
	}
	cmd.AddCommand(NewAutoCmd())
	cmd.AddCommand(NewInspectCmd())
	cmd.AddCommand(NewScanCmd())
	cmd.AddCommand(NewDiscoverCmd())
	cmd.AddCommand(NewMapCmd())
	cmd.AddCommand(NewSmartCmd())
	cmd.AddCommand(NewSwarmCmd())
	cmd.AddCommand(NewMagiswarmCmd())
	cmd.AddCommand(NewBlckswanCmd())

	cmd.AddCommand(NewOctoSwanCmd())

	return cmd
}
