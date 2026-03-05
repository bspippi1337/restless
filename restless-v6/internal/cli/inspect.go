package cli
import(
"fmt"
"github.com/spf13/cobra"
"github.com/bspippi1337/restless/internal/store"
)
func inspectCmd()*cobra.Command{
return &cobra.Command{
Use:"inspect <path>",
Run:func(cmd *cobra.Command,args []string){
api:=store.Last()
for _,e:=range api.Endpoints{
if e.Path==args[0]{
fmt.Println(e)
}
}
},
}
}
