package cli
import(
"fmt"
"io"
"net/http"
"github.com/spf13/cobra"
)
func callCmd()*cobra.Command{
return &cobra.Command{
Use:"call <method> <url>",
RunE:func(cmd *cobra.Command,args []string)error{
req,_:=http.NewRequest(args[0],args[1],nil)
res,err:=http.DefaultClient.Do(req)
if err!=nil{return err}
b,_:=io.ReadAll(res.Body)
fmt.Println(string(b))
return nil
},
}
}
