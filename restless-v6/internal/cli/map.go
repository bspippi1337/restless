package cli
import(
"fmt"
"github.com/spf13/cobra"
"github.com/bspippi1337/restless/internal/graph"
"github.com/bspippi1337/restless/internal/store"
)
func mapCmd()*cobra.Command{
return &cobra.Command{
Use:"map",
Run:func(cmd *cobra.Command,args []string){
api:=store.Last()
fmt.Println(graph.Render(api))
},
}
}
