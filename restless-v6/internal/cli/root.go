package cli
import(
"github.com/spf13/cobra"
)
func Root()*cobra.Command{
cmd:=&cobra.Command{
Use:"restless",
Short:"Explore and understand unknown REST APIs",
}
cmd.PersistentFlags().StringP("api","a","","api context")
cmd.PersistentFlags().StringP("cache","c","","cache directory")
cmd.AddCommand(discoverCmd())
cmd.AddCommand(mapCmd())
cmd.AddCommand(callCmd())
cmd.AddCommand(inspectCmd())
return cmd
}
