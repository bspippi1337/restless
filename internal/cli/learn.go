package cli

import (
"context"
"fmt"
"time"

"github.com/bspippi1337/restless/internal/httpx"
"github.com/bspippi1337/restless/internal/store"
"github.com/bspippi1337/restless/internal/util"
"github.com/spf13/cobra"
)

func NewLearnCmd() *cobra.Command {

return &cobra.Command{
Use:   "learn <url>",
Short: "Learn API and generate CLI shortcuts",
Args:  cobra.ExactArgs(1),

RunE: func(cmd *cobra.Command, args []string) error {

base := args[0]

api := &store.API{
BaseURL: base,
Endpoints: []store.Endpoint{
{Path: "/users"},
{Path: "/repos"},
},
}

cacheRoot,_ := cmd.Root().PersistentFlags().GetString("cache")
cacheRoot,_ = store.DefaultRoot(cacheRoot)

_,err := store.Write(cacheRoot,api)
if err != nil {
return err
}

fmt.Println("API learned:",base)
fmt.Println("Shortcuts available: users repos")

return nil
},
}
}

func NewDynamicCmd(name string,path string)*cobra.Command{

return &cobra.Command{
Use: name+" [id]",
Short: "Dynamic API shortcut",

RunE: func(cmd *cobra.Command,args []string)error{

apiName,_:=cmd.Root().PersistentFlags().GetString("api")
cacheRoot,_:=cmd.Root().PersistentFlags().GetString("cache")

cacheRoot,_=store.DefaultRoot(cacheRoot)

api,err:=store.Read(cacheRoot,apiName)
if err!=nil{
return err
}

endpoint:=path

if len(args)==1{
endpoint=util.JoinURL(path,args[0])
}

url:=util.JoinURL(api.BaseURL,endpoint)

client:=httpx.New()

ctx,_:=context.WithTimeout(context.Background(),10*time.Second)

res,err:=client.Do(ctx,"GET",url,nil)
if err!=nil{
return err
}

body,_:=httpx.ReadBody(res,2<<20)

fmt.Println(res.Status)
fmt.Println(string(body))

return nil
},
}
}
