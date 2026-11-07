import test from "node:test";
import assert from "node:assert/strict";
import {validateRequest,executeRequest,METHODS} from "../site/request/engine.mjs";
const base={url:"https://petstore3.swagger.io/api/v3/openapi.json",allowedHost:"petstore3.swagger.io",method:"GET",authorized:true,headersJson:"{}"};
test("Request Lab is restricted to three read-only HTTP methods",()=>{
  assert.deepEqual(METHODS,["GET","HEAD","OPTIONS"]);
  for(const method of ["POST","PATCH","PUT","DELETE","TRACE","CONNECT"])
    assert.throws(()=>validateRequest({...base,method}),/Only GET, HEAD and OPTIONS/);
});
test("Every request needs permission, HTTPS, and an exact allowed hostname",()=>{
  assert.throws(()=>validateRequest({...base,authorized:false}),/authorization confirmation/i);
  assert.throws(()=>validateRequest({...base,url:"http://petstore3.swagger.io/docs"}),/HTTPS/);
  assert.throws(()=>validateRequest({...base,allowedHost:"somewhere-else.example.org"}),/approved host/);
  assert.throws(()=>validateRequest({...base,url:"https://user:pass@petstore3.swagger.io/docs"}),/credentials/);
  assert.throws(()=>validateRequest({...base,url:"https://petstore3.swagger.io/docs#fragment"}),/fragments/);
});
test("Local addresses and IP literals are never approved",()=>{
  for(const host of ["127.0.0.1","localhost","localhost.localdomain","internal.local","api.internal","192.168.0.1","10.0.0.1","[::1]"]){
    assert.throws(()=>validateRequest({...base,allowedHost:host,url:"https://"+host+"/"}),/public DNS hostname/);
  }
});
test("Authorization headers require a second explicit confirmation",()=>{
  const conf={...base,headersJson:'{"Authorization":"Bearer test-token"}'};
  assert.throws(()=>validateRequest(conf),/separate credential confirmation/);
  const approved=validateRequest({...conf,allowAuthorization:true});
  assert.equal(approved.headers.get("Authorization"),"Bearer test-token");
});
test("Browser-managed headers and header injection are rejected",()=>{
  for(const name of ["Cookie","Host","Origin","Referer","Sec-Fetch-Mode","Proxy-Authorization"])
    assert.throws(()=>validateRequest({...base,headersJson:JSON.stringify({[name]:"abc"})}),/not permitted/);
  assert.throws(()=>validateRequest({...base,headersJson:'{"Accept":"text/plain\r\nX: evil"}'}),/Invalid headers JSON/);
  assert.throws(()=>validateRequest({...base,headersJson:'{"x":7}'}),/strings/);
});
test("Success records one request, SHA-256, elapsed time and browser-safe configuration",async()=>{
  const calls=[];
  const mock=async(url,init)=>{
    calls.push({url,init});
    return new Response('{"ok":true}',{status:200,headers:{"content-type":"application/json","x-what":"demo"}});
  };
  const result=await executeRequest({...base,headersJson:'{"Accept":"application/json"}'},mock);
  assert.equal(calls.length,1);
  assert.equal(calls[0].init.method,"GET");
  assert.equal(calls[0].init.credentials,"omit");
  assert.equal(calls[0].init.redirect,"error");
  assert.equal(calls[0].init.mode,"cors");
  assert.equal(calls[0].init.referrerPolicy,"no-referrer");
  assert.equal(result.networkRequests,1);
  assert.equal(result.bodyFormat,"formatted-json");
  assert.match(result.responseBody,/"ok": true/);
  assert.equal(result.sha256.length,64);
  assert.equal(result.responseHeaders["x-what"],"demo");
});
test("HEAD response without a body is supported",async()=>{
  const out=await executeRequest({...base,method:"HEAD"},async()=>new Response(null,{status:204}));
  assert.equal(out.networkRequests,1);
  assert.equal(out.bytesRead,0);
  assert.equal(out.responseBody,"");
});
test("Oversized content-length is rejected before reading",async()=>{
  await assert.rejects(
    executeRequest(base,async()=>new Response("x",{headers:{"content-length":"3000000"}})),
    /2 MB/
  );
});
test("Oversized streamed body is aborted",async()=>{
  const large="x".repeat(2_000_001);
  await assert.rejects(executeRequest(base,async()=>new Response(large)),/2 MB/);
});
test("Network/CORS failures are not hidden or retried",async()=>{
  let count=0;
  await assert.rejects(executeRequest(base,async()=>{count++;throw new TypeError("Failed to fetch");}),/CORS/);
  assert.equal(count,1);
});
test("Timeout range is bounded",()=>{
  assert.throws(()=>validateRequest({...base,timeoutMs:999}),/1–15/);
  assert.throws(()=>validateRequest({...base,timeoutMs:15001}),/1–15/);
});


test("Evidence reports redact sensitive URL query values without changing the one request sent",async()=>{
 const url="https://petstore3.swagger.io/api/v3/pets?access_token=PRIVATE-SAMPLE&language=en";
 let called="";
 const report=await executeRequest({...base,url},async target=>{called=target;return new Response("OK");});
 assert.equal(called,url);
 assert.ok(!report.url.includes("PRIVATE-SAMPLE"));
 assert.match(report.url,/REDACTED/);
 assert.match(report.url,/language=en/);
});
