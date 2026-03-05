package cli

import (
"fmt"
"strings"

"github.com/spf13/cobra"

"github.com/bspippi1337/restless/internal/store"
)

func NewMapCmd() *cobra.Command {

return &cobra.Command{
Use:   "map",
Short: "Render API endpoint topology",

RunE: func(cmd *cobra.Command, args []string) error {

cacheRoot,_ := cmd.Root().PersistentFlags().GetString("cache")
apiName,_ := cmd.Root().PersistentFlags().GetString("api")

cacheRoot,_ = store.DefaultRoot(cacheRoot)

api,err := store.Read(cacheRoot,apiName)
if err != nil {
return err
}

if api == nil || len(api.Endpoints) == 0 {
fmt.Println("No endpoints discovered yet.")
fmt.Println("Run: restless learn <url>")
return nil
}

fmt.Println("API MAP")
fmt.Println()

for i,e := range api.Endpoints {

path := strings.TrimPrefix(e.Path,"/")

if path == "" {
fmt.Println("/")
continue
}

prefix := "├──"
if i == len(api.Endpoints)-1 {
prefix = "└──"
}

fmt.Printf("%s %s\n",prefix,path)

}

return nil
},
}
}
