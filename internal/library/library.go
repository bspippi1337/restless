// Package library turns a single domain into a source-backed, bounded API
// documentation library, never into a fuzzing or endpoint-guessing target.
package library

import (
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "net"
 "net/http"
 "net/url"
 "os"
 "path/filepath"
 "regexp"
 "strings"
 "time"

 "github.com/bspippi1337/restless/internal/botanical"
)

const DefaultAgent = "RESTLESS42-Library/1 (+https://github.com/bspippi1337/restless)"
const maxPages = 25

var hostLabel=regexp.MustCompile("^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$")

type Target struct {
 Host string `json:"host"`
 SeedURL string `json:"seed_url"`
}

// Normalize only accepts an explicit public DNS name or an explicit HTTPS
// document URL. No endpoint dictionaries, ports or host enumeration.
func Normalize(input string)(Target,error){
 var empty Target
 v:=strings.TrimSpace(input)
 if v==""||len(v)>2048{return empty,errors.New("enter one domain or HTTPS documentation URL")}
 if !strings.Contains(v,"://"){v="https://"+v}
 u,err:=url.Parse(v)
 if err!=nil||u==nil||u.Scheme!="https"||u.Host==""{return empty,errors.New("HTTPS domain or URL required")}
 if u.User!=nil||u.RawQuery!=""||u.Fragment!=""||u.Port()!=""{return empty,errors.New("credentials, query strings, fragments and custom ports are not allowed")}
 host:=strings.ToLower(u.Hostname())
 if net.ParseIP(host)!=nil||len(host)>253||strings.HasSuffix(host,"."){return empty,errors.New("use a public DNS hostname, not an IP literal")}
 if !strings.Contains(host,"."){return empty,errors.New("public hostname must contain a dot")}
 for _,part:=range strings.Split(host,"."){
  if !hostLabel.MatchString(part){return empty,errors.New("domain contains invalid DNS labels")}
 }
 for _,suffix:=range []string{".local",".localhost",".internal",".test",".invalid",".example",".onion",".home",".lan"}{
  if strings.HasSuffix(host,suffix){return empty,errors.New("reserved/private domains are not accepted")}
 }
 if strings.ContainsAny(u.EscapedPath(),"\r\n"){return empty,errors.New("unsafe URL path")}
 if u.Path==""{u.Path="/"}
 u.Host=host
 u.RawPath=""
 return Target{Host:host,SeedURL:u.String()},nil
}

type Options struct {
 Input string
 OutputDir string
 MaxPages int
 MaxDepth int
 Delay time.Duration
 Execute bool
 Authorized bool
 Agent string
 OnDocument func(botanical.Record)
 // Client is ONLY for controlled test transports. Production defaults to
 // Botanical's public-IP-only dialer, no redirects and no implicit proxies.
 Client *http.Client
}

type Summary struct {
 SchemaVersion int `json:"schema_version"`
 Target Target `json:"target"`
 Mode string `json:"mode"`
 Authorized bool `json:"authorized"`
 NetworkRequests string `json:"network_requests"`
 OutputDir string `json:"output_dir"`
 MaxPages int `json:"max_pages"`
 MaxDepth int `json:"max_depth"`
 DelaySeconds int `json:"delay_seconds"`
 GeneratedAt time.Time `json:"generated_at"`
 DocumentCount int `json:"document_count"`
 OperationCount int `json:"operation_count"`
 Methods map[string]int `json:"documented_methods"`
 Documents []botanical.Document `json:"documents,omitempty"`
 Operations []botanical.Operation `json:"operations,omitempty"`
 Records []botanical.Record `json:"records,omitempty"`
 Warnings []string `json:"warnings"`
}

func defaults(o Options)Options{
 if o.MaxPages==0{o.MaxPages=12}
 if o.Delay==0{o.Delay=2*time.Second}
 if o.Agent==""{o.Agent=DefaultAgent}
 if o.OutputDir==""{o.OutputDir=filepath.Join(".restless","libraries")}
 return o
}

func Preview(options Options)(Summary,error){
 o:=defaults(options)
 t,err:=Normalize(o.Input);if err!=nil{return Summary{},err}
 if o.MaxPages<1||o.MaxPages>maxPages{return Summary{},fmt.Errorf("max pages must be 1..%d",maxPages)}
 if o.MaxDepth<0||o.MaxDepth>3{return Summary{},errors.New("max depth must be 0..3")}
 if o.Delay<2*time.Second{o.Delay=2*time.Second}
 if strings.TrimSpace(o.Agent)==""{return Summary{},errors.New("identifiable user-agent required")}
 return Summary{
  SchemaVersion:1,Target:t,Mode:"DRY_RUN",Authorized:false,NetworkRequests:"0",
  OutputDir:filepath.Join(o.OutputDir,t.Host),MaxPages:o.MaxPages,
  MaxDepth:o.MaxDepth,DelaySeconds:int(o.Delay.Seconds()),GeneratedAt:time.Now().UTC(),
  Methods:map[string]int{},
  Warnings:[]string{
   "Preview only. No DNS resolution, robots.txt fetch or HTTP request.",
   "Network access requires explicit execution AND permission.",
   "Only linked documentation and declared OpenAPI contracts are discovered. API endpoints are never guessed.",
   "Documented methods are inventory entries; no API operations are executed.",
  },
 },nil
}

// Build uses the existing audited Botanical crawler, validates SHA256 archive
// bytes, and writes a machine-readable library.json plus catalog.json.
func Build(ctx context.Context,options Options)(Summary,error){
 s,err:=Preview(options);if err!=nil{return s,err}
 if !options.Execute{return s,nil}
 if !options.Authorized{return s,errors.New("refusing network requests without explicit authorization")}
 o:=defaults(options)
 if o.Delay<2*time.Second{o.Delay=2*time.Second}
 records,err:=botanical.Crawl(ctx,botanical.Config{
  SeedURL:s.Target.SeedURL,AllowedHost:s.Target.Host,Agent:o.Agent,
  OutputDir:s.OutputDir,MaxPages:s.MaxPages,MaxDepth:s.MaxDepth,
  Delay:o.Delay,Authorized:true,Client:o.Client,OnDocument:o.OnDocument,
 })
 s.Mode="ARCHIVE"
 s.Authorized=true
 s.NetworkRequests="robots.txt plus up to the configured document request budget"
 s.Records=records
 if err!=nil{return s,fmt.Errorf("documentation crawl incomplete: %w",err)}
 if len(records)==0{return s,errors.New("no documentation archived; supply an explicitly known documentation URL or check robots.txt")}
 c,err:=botanical.BuildCatalog(s.OutputDir)
 if err!=nil{return s,fmt.Errorf("cannot verify saved archive: %w",err)}
 if err=botanical.SaveCatalog(s.OutputDir,c);err!=nil{return s,err}
 s.DocumentCount=len(c.Documents)
 s.OperationCount=len(c.Operations)
 s.Documents=c.Documents;s.Operations=c.Operations
 for _,op:=range c.Operations{s.Methods[op.Method]++}
 s.Warnings=[]string{
  "Only declared OpenAPI HTTP operations were indexed, never executed.",
  "A documented API operation is not proof of a deployed endpoint.",
  "PDF sources are archived and hashed but not interpreted as API operations.",
 }
 s.GeneratedAt=time.Now().UTC()
 data,err:=json.MarshalIndent(s,"","  ");if err!=nil{return s,err}
 tmp,err:=os.CreateTemp(s.OutputDir,".library-*");if err!=nil{return s,err}
 defer os.Remove(tmp.Name())
 if err=tmp.Chmod(0600);err!=nil{tmp.Close();return s,err}
 if _,err=tmp.Write(append(data,'\n'));err!=nil{tmp.Close();return s,err}
 if err=tmp.Close();err!=nil{return s,err}
 if err=os.Rename(tmp.Name(),filepath.Join(s.OutputDir,"library.json"));err!=nil{return s,err}
 return s,nil
}
