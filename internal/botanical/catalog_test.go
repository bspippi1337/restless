package botanical

import (
 "crypto/sha256"
 "encoding/hex"
 "encoding/json"
 "os"
 "path/filepath"
 "strings"
 "testing"
)

func fixture(t *testing.T, content, mediaType string) string {
 t.Helper()
 dir:=t.TempDir()
 if err:=os.MkdirAll(filepath.Join(dir,"sha256"),0700);err!=nil{t.Fatal(err)}
 sum:=sha256.Sum256([]byte(content))
 hash:=hex.EncodeToString(sum[:])
 if err:=os.WriteFile(filepath.Join(dir,"sha256",hash),[]byte(content),0600);err!=nil{t.Fatal(err)}
 record:=Record{URL:"https://example.org/openapi.json",Source:"seed",SHA256:hash,ArchivePath:filepath.Join("sha256",hash),ContentType:mediaType}
 b,err:=json.Marshal(record)
 if err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(filepath.Join(dir,"index.jsonl"),append(b,'\n'),0600);err!=nil{t.Fatal(err)}
 return dir
}

func TestCatalogAndSourceBoundPlan(t *testing.T) {
 spec:=`{"openapi":"3.0.4","info":{"title":"Petstore docs"},"security":[{"api_key":[]}],"paths":{"/pets/{petId}":{"parameters":[{"name":"petId","in":"path","required":true,"schema":{"type":"integer"}}],"get":{"operationId":"findPetById","summary":"Find pet by ID","tags":["pets"]},"delete":{"operationId":"deletePet","summary":"Delete pet","security":[]}}}}`
 dir:=fixture(t,spec,"application/json")
 c,err:=BuildCatalog(dir)
 if err!=nil{t.Fatal(err)}
 if len(c.Documents)!=1 || len(c.Operations)!=2 {t.Fatalf("docs=%d operations=%d",len(c.Documents),len(c.Operations))}
 if c.Documents[0].Kind!="openapi-3" {t.Fatalf("unexpected kind %s",c.Documents[0].Kind)}
 if err:=SaveCatalog(dir,c);err!=nil{t.Fatal(err)}
 loaded,err:=LoadCatalog(dir)
 if err!=nil{t.Fatal(err)}
 hits,err:=RankOperations(loaded,"find pet by id",10)
 if err!=nil{t.Fatal(err)}
 if len(hits)==0 || hits[0].Operation.Method!="GET" {t.Fatalf("unexpected ranked operations %+v",hits)}
 plan,err:=BuildProbePlan(loaded,"find pet by id",3)
 if err!=nil{t.Fatal(err)}
 if plan.NetworkRequests!=0 || plan.Mode!="DOCUMENTATION_ONLY_NO_EXECUTION" {t.Fatalf("unsafe plan %+v",plan)}
 top:=plan.Candidates[0]
 if !top.RequiresManualReview || len(top.Parameters)!=1 || !top.OperationAuth() {t.Fatalf("missing auth/params in plan: %+v",top)}
 if top.SourceURL!="https://example.org/openapi.json" || top.SourceSHA256=="" {t.Fatalf("missing provenance %+v",top)}
 if len(top.ReviewReasons)==0 {t.Fatal("expected review reasons")}
}

func (s ProbeStep) OperationAuth() bool { return len(s.Security)>0 }

func TestTamperedArchiveRejected(t *testing.T) {
 dir:=fixture(t,`{"openapi":"3.0.4","paths":{}}`,"application/json")
 archive,err:=filepath.Glob(filepath.Join(dir,"sha256","*"))
 if err!=nil || len(archive)!=1 {t.Fatal(err,archive)}
 if err:=os.WriteFile(archive[0],[]byte("modified"),0600);err!=nil{t.Fatal(err)}
 _,err=BuildCatalog(dir)
 if err==nil||!strings.Contains(err.Error(),"SHA256 mismatch"){t.Fatalf("unexpected result: %v",err)}
}

func TestUnstructuredDocsDoNotInventEndpoints(t *testing.T) {
 dir:=fixture(t,"GET /secrets\nPOST /unknown\n","text/plain")
 c,err:=BuildCatalog(dir)
 if err!=nil{t.Fatal(err)}
 if len(c.Operations)!=0{t.Fatalf("fabricated %+v",c.Operations)}
 plan,err:=BuildProbePlan(c,"get secrets",5)
 if err!=nil{t.Fatal(err)}
 if len(plan.Candidates)!=0||plan.NetworkRequests!=0{t.Fatalf("bad plan %+v",plan)}
}

func TestUnresolvedRefsRequireReview(t *testing.T) {
 spec:=`{"openapi":"3.0.4","paths":{"/pets":{"post":{"operationId":"createPet","summary":"Create pet","parameters":[{"$ref":"https://attacker.example/params"}]}}}}`
 dir:=fixture(t,spec,"application/json")
 c,err:=BuildCatalog(dir)
 if err!=nil{t.Fatal(err)}
 if len(c.Operations)!=1||!c.Operations[0].UnresolvedRefs{t.Fatalf("bad reference detection %+v",c.Operations)}
 p,err:=BuildProbePlan(c,"create pet",1)
 if err!=nil{t.Fatal(err)}
 if len(p.Candidates)!=1||!p.Candidates[0].RequiresManualReview||p.Candidates[0].Confidence!="partial"{t.Fatalf("bad review %+v",p)}
}

func TestBadManifestPathRejected(t *testing.T) {
 dir:=fixture(t,"some text","text/plain")
 index:=filepath.Join(dir,"index.jsonl")
 b,err:=os.ReadFile(index)
 if err!=nil{t.Fatal(err)}
 var r Record
 if err:=json.Unmarshal(b,&r);err!=nil{t.Fatal(err)}
 r.ArchivePath="../../credentials"
 changed,err:=json.Marshal(r)
 if err!=nil{t.Fatal(err)}
 if err:=os.WriteFile(index,append(changed,'\n'),0600);err!=nil{t.Fatal(err)}
 if _,err:=BuildCatalog(dir);err==nil{t.Fatal("accepted path traversal")}
}
