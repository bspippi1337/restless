package botanical

import (
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "net/http"
 "os"
 "path/filepath"
 "testing"
 "time"
)
func TestRequestAndResponseSamplesAreExplicitAndBounded(t *testing.T){
 raw:=rawRequestBody{Required:true,Content:map[string]rawMediaType{
  "application/json":{
   Schema:rawSchema{Type:"object",Properties:map[string]rawSchema{
    "name":{Type:"string",Example:"Pixel"},
    "count":{Type:"integer",Default:4},
    "password":{Type:"string",Example:"NEVER-EXPOSE-THIS"},
    "scope":{Type:"string",Enum:[]any{"read","write"}},
   }},
  },
 }}
 result:=summarizeRequest(raw)
 if result==nil||!result.Required||result.ContentType!="application/json"{t.Fatalf("wrong request %#v",result)}
 fields,ok:=result.Example.(map[string]any)
 if !ok{t.Fatalf("expected example object got %T",result.Example)}
 if fields["name"]!="Pixel"||fields["count"]!=4||fields["scope"]!="read"{t.Fatalf("invalid examples: %#v",fields)}
 if fields["password"]=="NEVER-EXPOSE-THIS"{t.Fatal("secret example was leaked")}
 rawResponses:=map[string]rawResponse{
  "200":{Description:"Successful response",Content:map[string]rawMediaType{
   "application/json":{Schema:rawSchema{Type:"object",Properties:map[string]rawSchema{"ok":{Type:"boolean"}}}},
  }},
  "404":{Description:"Not found"},
 }
 responses:=summarizeResponses(rawResponses)
 if len(responses)!=2||responses[0].Status!="200"||responses[1].Status!="404"{t.Fatalf("wrong response sort %#v",responses)}
 if responses[0].ContentType!="application/json"{t.Fatal("missing response media type")}
}
func TestUnresolvedReferenceRemainsUnevaluated(t *testing.T){
 r:=summarizeRequest(rawRequestBody{Ref:"#/components/requestBodies/Order"})
 if r==nil||r.UnresolvedRef!="#/components/requestBodies/Order"{t.Fatalf("%#v",r)}
 if r.Example!=nil{t.Fatalf("do not guess referenced example %#v",r.Example)}
}
func TestSchemaDepthAndPropertyCountBounded(t *testing.T){
 props:=map[string]rawSchema{}
 for i:=0;i<70;i++{props[string(rune('A'+i))]=rawSchema{Type:"string"}}
 sample:=schemaExample(rawSchema{Type:"object",Properties:props},"root",0)
 obj:=sample.(map[string]any)
 if len(obj)>21{t.Fatalf("too many properties %d",len(obj))}
 nested:=rawSchema{Type:"object",Properties:map[string]rawSchema{"inner":{Type:"object",Properties:map[string]rawSchema{"depth":{Type:"string"}}}}}
 if schemaExample(nested,"root",4)!="<nested schema>"{t.Fatal("depth cap not enforced")}
}

func TestMediaExamplesRedactNamedSecretFields(t *testing.T) {
  raw := rawRequestBody{Content: map[string]rawMediaType{
    "application/json":{Example: map[string]any{"name":"Widget","api_key":"DO-NOT-EXPORT","nested":map[string]any{"password":"private-value"}}},
  }}
  example := summarizeRequest(raw)
  if example == nil {t.Fatal("missing request sample")}
  doc := example.Example.(map[string]any)
  if doc["api_key"] == "DO-NOT-EXPORT" {t.Fatal("API key leaked through media example")}
  nested := doc["nested"].(map[string]any)
  if nested["password"] == "private-value" {t.Fatal("nested password leaked through media example")}
}

func TestCatalogPreservesOpenAPIBodiesResponsesAndEvidence(t *testing.T) {
  directory:=t.TempDir()
  text:=[]byte(`{"openapi":"3.0.3","info":{"title":"Pets"},"paths":{"/pets":{"post":{"operationId":"createPet","summary":"Create pet","requestBody":{"required":true,"content":{"application/json":{"schema":{"type":"object","properties":{"name":{"type":"string","example":"Pixel"},"password":{"type":"string","example":"topsecret"}}}}}},"responses":{"201":{"description":"Created","content":{"application/json":{"schema":{"type":"object","properties":{"id":{"type":"integer","example":42}}}}}},"400":{"description":"Bad request"}}}}}}`)
  digest:=sha256.Sum256(text)
  hash:=hex.EncodeToString(digest[:])
  if err:=os.Mkdir(filepath.Join(directory,"sha256"),0700);err!=nil{t.Fatal(err)}
  if err:=os.WriteFile(filepath.Join(directory,"sha256",hash),text,0600);err!=nil{t.Fatal(err)}
  record:=Record{
    URL:"https://docs.example.org/openapi.json",Status:http.StatusOK,
    ContentType:"application/json",SHA256:hash,
    ArchivePath:filepath.Join("sha256",hash),Bytes:len(text),FetchedAt:time.Now().UTC(),
  }
  encoded,err:=json.Marshal(record);if err!=nil{t.Fatal(err)}
  if err=os.WriteFile(filepath.Join(directory,"index.jsonl"),append(encoded,'
'),0600);err!=nil{t.Fatal(err)}
  catalog,err:=BuildCatalog(directory);if err!=nil{t.Fatal(err)}
  if len(catalog.Operations)!=1{t.Fatalf("expected one explicit operation, got %d",len(catalog.Operations))}
  op:=catalog.Operations[0]
  if op.Method!="POST"||op.Path!="/pets"||op.SourceSHA256!=hash{t.Fatalf("incorrect provenance: %+v",op)}
  if op.Request==nil||op.Request.ContentType!="application/json"{t.Fatalf("missing request source body: %#v",op.Request)}
  fields,ok:=op.Request.Example.(map[string]any)
  if !ok||fields["name"]!="Pixel"{t.Fatalf("incorrect request schema example: %#v",op.Request.Example)}
  if fields["password"]=="topsecret"{t.Fatal("secret exposed in catalog sample")}
  if len(op.Responses)!=2||op.Responses[0].Status!="201"||op.Responses[1].Status!="400"{t.Fatalf("incorrect response inventory: %+v",op.Responses)}
  response,ok:=op.Responses[0].Example.(map[string]any)
  if !ok||response["id"]!=42{t.Fatalf("wrong documented response example: %#v",op.Responses[0].Example)}
  // The JSON library export must retain the same statuses, source hash and sample.
  data,err:=json.Marshal(op);if err!=nil{t.Fatal(err)}
  var decoded Operation
  if err=json.Unmarshal(data,&decoded);err!=nil{t.Fatal(err)}
  if decoded.Request==nil||len(decoded.Responses)!=2||decoded.SourceSHA256!=hash{t.Fatal("JSON roundtrip lost catalog evidence")}
}
