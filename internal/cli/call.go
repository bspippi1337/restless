package cli

import (
"context"
"fmt"
"strings"
"time"

"github.com/bspippi1337/restless/internal/httpx"
"github.com/bspippi1337/restless/internal/store"
"github.com/bspippi1337/restless/internal/util"
"github.com/spf13/cobra"
)

func NewCallCmd() *cobra.Command {

var timeout time.Duration

cmd := &cobra.Command{
Use:   "call <METHOD> <PATH>",
Short: "Call API endpoint",
Args:  cobra.ExactArgs(2),

RunE: func(cmd *cobra.Command, args []string) error {

method := strings.ToUpper(args[0])
path := args[1]

apiName,_ := cmd.Root().PersistentFlags().GetString("api")
cacheRoot,_ := cmd.Root().PersistentFlags().GetString("cache")

cacheRoot,_ = store.DefaultRoot(cacheRoot)

api,err := store.Read(cacheRoot,apiName)
if err != nil {
return err
}

url := util.JoinURL(api.BaseURL,path)

ctx,_ := context.WithTimeout(context.Background(),timeout)

client := httpx.New()

res,err := client.Do(ctx,method,url,nil)
if err != nil {
return err
}

body,_ := httpx.ReadBody(res,2<<20)

fmt.Println(method,url)
fmt.Println(res.Status)
fmt.Println(string(body))

return nil
},
}

cmd.Flags().DurationVarP(&timeout,"timeout","t",10*time.Second,"timeout")

return cmd
}
