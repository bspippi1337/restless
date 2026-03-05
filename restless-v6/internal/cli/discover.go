package cli
import(
"fmt"
"github.com/spf13/cobra"
"github.com/bspippi1337/restless/internal/discovery"
)
func discoverCmd()*cobra.Command{
return &cobra.Command{
Use:"discover <url>",
RunE:func(cmd *cobra.Command,args []string)error{
res,err:=discovery.Run(args[0])
if err!=nil{return err}
fmt.Printf("discovered %d endpoints\n",len(res.Endpoints))
for _,e:=range res.Endpoints{
fmt.Println(e.Path,e.Methods)
}
return nil
},
}
}
