package discovery
import(
"encoding/json"
"io"
"net/http"
"regexp"
"strings"
"sync"
"github.com/bspippi1337/restless/internal/store"
"github.com/bspippi1337/restless/internal/util"
)

type Result struct{
Endpoints []store.Endpoint
}

var openapiPaths=[]string{
"/swagger.json",
"/openapi.json",
"/v3/api-docs",
"/api-docs",
}

var linkRe=regexp.MustCompile(`href="([^"]+)"`)

func Run(base string)(Result,error){

base=util.Normalize(base)

var endpoints []store.Endpoint

for _,p:=range openapiPaths{

url:=base+p

res,err:=http.Get(url)
if err!=nil{continue}

body,_:=io.ReadAll(res.Body)

if strings.Contains(string(body),"paths"){

var doc map[string]interface{}

json.Unmarshal(body,&doc)

paths:=doc["paths"]

if m,ok:=paths.(map[string]interface{});ok{

for k,v:=range m{

methods:=[]string{}

if mm,ok:=v.(map[string]interface{});ok{

for mk:=range mm{
methods=append(methods,strings.ToUpper(mk))
}

}

endpoints=append(endpoints,store.Endpoint{
Path:k,
Methods:methods,
})

}

}

}

}

if len(endpoints)>0{
return Result{Endpoints:endpoints},nil
}

res,err:=http.Get(base)
if err!=nil{return Result{},err}

body,_:=io.ReadAll(res.Body)

matches:=linkRe.FindAllSubmatch(body,-1)

var wg sync.WaitGroup
var mu sync.Mutex

for _,m:=range matches{

path:=string(m[1])

if !strings.HasPrefix(path,"/"){continue}

wg.Add(1)

go func(p string){

defer wg.Done()

url:=base+p

r,err:=http.Get(url)

if err!=nil{return}

if r.StatusCode<400{

mu.Lock()

endpoints=append(endpoints,store.Endpoint{
Path:p,
Methods:[]string{"GET"},
})

mu.Unlock()

}

}(path)

}

wg.Wait()

store.Save(store.API{
Base:base,
Endpoints:endpoints,
})

return Result{Endpoints:endpoints},nil
}
