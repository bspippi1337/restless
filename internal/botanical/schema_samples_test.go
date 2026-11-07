package botanical

import (
 "testing"
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
