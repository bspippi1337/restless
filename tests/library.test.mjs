import test from "node:test";
import assert from "node:assert/strict";
import {validateDomain,readAPI,makeTree,makeExample,toPostman,toMermaid} from "../site/library/engine.mjs";
const spec={
 openapi:"3.0.3",info:{title:"Inventory API",version:"7.1"},
 servers:[{url:"https://api.inventory.example.net/v1"}],
 paths:{
  "/items":{get:{operationId:"listItems",summary:"List inventory",parameters:[{name:"limit",in:"query",schema:{type:"integer",example:10}}]},
    post:{operationId:"createItem",summary:"Create item",requestBody:{content:{"application/json":{schema:{type:"object",properties:{name:{type:"string",example:"widget"},stock:{type:"integer",default:0}}}}}}}},
  "/items/{id}":{parameters:[{name:"id",in:"path",required:true,schema:{type:"string",format:"uuid"}}],get:{operationId:"findItem",summary:"Find item"}},
  "/items/{id}/images":{get:{operationId:"listImages",summary:"List images"}}
 }
};
test("Domain parser normalizes public HTTPS but never invents endpoint paths",()=>{
 assert.deepEqual(validateDomain("api.inventory.example.net"),{host:"api.inventory.example.net",url:"https://api.inventory.example.net/",explicitPath:false});
 assert.deepEqual(validateDomain("https://docs.api.example.net/openapi.json"),{host:"docs.api.example.net",url:"https://docs.api.example.net/openapi.json",explicitPath:true});
});
test("Reject IPs, reserved hosts, credentials, query and non-HTTPS",()=>{
 const bad=["http://api.example.net/","localhost","10.0.0.1","https://192.168.1.1",
   "https://[::1]/","api.local","api.internal","https://x.example.invalid",
   "https://user:secret@api.example.net","https://api.example.net/?key=secret",
   "https://api.example.net/#frag","https://api.example.net:443/"];
 for(const input of bad)assert.throws(()=>validateDomain(input),undefined,input);
});
test("Parse documented operations and create a reproducible tree of explicit paths",()=>{
 const model=readAPI(spec),tree=makeTree(model);
 assert.equal(model.networkRequests,0);
 assert.equal(model.stats.operations,4);
 assert.equal(model.stats.paths,3);
 assert.equal(model.stats.methods.GET,3);
 assert.equal(model.stats.methods.POST,1);
 assert.equal(tree.nodes.length,4);
 assert.equal(tree.edges.length,3);
 assert.ok(tree.nodes.some(x=>x.path==="/items/{id}/images"&&x.operations[0].operationId==="listImages"));
 assert.equal(model.server,"https://api.inventory.example.net/v1");
});
test("Show type-based input hints and example JSON without sending a request",()=>{
 const model=readAPI(spec);
 const get=makeExample(model,model.operations.find(x=>x.operationId==="findItem"));
 assert.equal(get.method,"GET");
 assert.match(get.displayURL,/items\/\{\{id\}\}/);
 assert.equal(get.parameters[0].example,"00000000-0000-0000-0000-000000000000");
 assert.equal(get.reviewRequired,true);
 const post=makeExample(model,model.operations.find(x=>x.method==="POST"));
 assert.deepEqual({...post.body},{name:"widget",stock:0});
 assert.match(post.curl,/curl --request POST/);
 assert.match(post.curl,/PREVIEW|url/);
 assert.equal(post.source.sha256,"");
});
test("Go crawler library.json imports provenance and documented methods",()=>{
 const sample={
  schema_version:1,target:{host:"api.inventory.example.net",seed_url:"https://api.inventory.example.net/docs"},
  generated_at:"2026-11-07T19:00:00Z",
  documents:[{url:"https://api.inventory.example.net/docs/openapi.json",sha256:"f".repeat(64),kind:"openapi-3"}],
  operations:[{method:"GET",path:"/inventory/{sku}",operation_id:"inventoryBySKU",parameters:[{name:"sku",in:"path",required:true,schema:{type:"string"}}],source_url:"https://api.inventory.example.net/docs/openapi.json",source_sha256:"f".repeat(64)}]
 };
 const model=readAPI(JSON.stringify(sample));
 assert.equal(model.kind,"RESTLESS_LIBRARY");
 assert.equal(model.operations[0].sourceSHA256,"f".repeat(64));
 assert.equal(model.stats.documents,1);
 assert.equal(makeExample(model,model.operations[0]).source.url,"https://api.inventory.example.net/docs/openapi.json");
 assert.equal(model.networkRequests,0);
});
test("Postman export only contains documented operations and safely disabled query examples",()=>{
 const model=readAPI(spec),collection=toPostman(model);
 assert.equal(collection.item.length,4);
 assert.equal(collection.info.schema,"https://schema.getpostman.com/json/collection/v2.1.0/collection.json");
 assert.equal(collection.variable[0].key,"baseUrl");
 const get=collection.item.find(i=>i.request.url.raw.endsWith("/items")&&i.request.method==="GET");
 assert.equal(get.request.url.query[0].disabled,true);
 const mutation=collection.item.find(i=>i.request.method==="POST");
 assert.equal(mutation.request.body.mode,"raw");
 assert.match(collection.info.description,/No endpoints guessed/);
});
test("Mermaid graph contains real path nodes, no guessed endpoints",()=>{
 const m=toMermaid(readAPI(spec));
 assert.match(m,/flowchart LR/);
 assert.match(m,/items/);
 assert.match(m,/images/);
 assert.match(m,/-->/);
 assert.ok(!m.includes("/admin"));
});
test("Unsupported formats and unsafe payload sizes fail closed",()=>{
 assert.throws(()=>readAPI('{"swagger":"2.0","paths":{}}'),/Expected OpenAPI 3/);
 assert.throws(()=>readAPI('not json'),/Invalid JSON/);
 assert.throws(()=>readAPI("x".repeat(2_000_001)),/2 million/);
});
test("Native catalog.json supports source traceability",()=>{
 const doc={version:1,documents:[{url:"https://api.example.net/openapi.json",sha256:"a".repeat(64)}],
  operations:[{method:"PATCH",path:"/records/{id}",source_url:"https://api.example.net/openapi.json",source_sha256:"a".repeat(64)}]};
 const m=readAPI(doc);
 assert.equal(m.kind,"RESTLESS_CATALOG");
 assert.equal(m.operations[0].method,"PATCH");
 assert.equal(m.operations[0].sourceSHA256,"a".repeat(64));
});
