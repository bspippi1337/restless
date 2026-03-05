package cli

import (
"fmt"
"os"

"github.com/spf13/cobra"
)

var (
version = "v6.0.0"
commit  = "dev"
date    = "unknown"
)

func NewRootCmd() *cobra.Command {

cmd := &cobra.Command{
Use:   "restless",
Short: "Explore and understand unknown REST APIs",
}

cmd.PersistentFlags().StringP("api", "a", "", "API context")
cmd.PersistentFlags().StringP("cache", "c", "", "cache directory")

cmd.AddCommand(NewDiscoverCmd())
cmd.AddCommand(NewMapCmd())
cmd.AddCommand(NewCallCmd())
cmd.AddCommand(NewInspectCmd())

cmd.Run = func(cmd *cobra.Command, args []string) {
cmd.Help()
}

cmd.SetVersionTemplate("restless {{.Version}}\n")
cmd.Version = fmt.Sprintf("%s (%s %s)", version, commit, date)

return cmd
}

func Execute() {
root := NewRootCmd()
if err := root.Execute(); err != nil {
fmt.Println(err)
os.Exit(1)
}
}
