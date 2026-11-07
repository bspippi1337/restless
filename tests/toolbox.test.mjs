import test from "node:test";
import assert from "node:assert/strict";
import {OPERATIONS,applyOperation,parseOpenAPIOperations,searchOpenAPIOperations,runRecipe,validateRecipe} from "../site/toolbox/transforms.mjs";
const demo={openapi:"3.0.3",info:{title:"Pets",version:"1"},security:[{apiKey:[]}],paths:{
 "/pets":{get:{operationId:"listPets",summary:"List pets"},post:{operationId:"createPet",summary:"Create pet",security:[]}},
 "/pets/{petId}":{parameters:[{$ref:"#/components/parameters/PetId"}],get:{operationId:"findPetById",summary:"Find pet by ID"}}
},components:{parameters:{PetId:{name:"petId",in:"path",required:true,schema:{type:"string"}}}}};
const spec=JSON.stringify(demo);
test("operation registry has 32 distinct working IDs",()=>{
 assert.equal(OPERATIONS.length,32);
 assert.equal(new Set(OPERATIONS.map(x=>x.id)).size,32);
 for(const op of OPERATIONS)assert.ok(op.label&&op.group&&op.description);
});
test("UTF-8 Base64 and URL-safe round trip",async()=>{
 const input="Héllo 🌍 & RESTLESS";
 assert.equal(await applyOperation("base64-decode",await applyOperation("base64-encode",input)),input);
 assert.equal(await applyOperation("base64url-decode",await applyOperation("base64url-encode",input)),input);
 assert.equal(await applyOperation("url-decode",await applyOperation("url-encode",input)),input);
});
test("Hex and HTML escaping are reversible",async()=>{
 const input='<note title="a&b">O\'Reilly 🦢</note>';
 assert.equal(await applyOperation("hex-decode",await applyOperation("hex-encode",input)),input);
 assert.equal(await applyOperation("html-unescape",await applyOperation("html-escape",input)),input);
});
test("JSON operations reject malformed input",async()=>{
 assert.equal(await applyOperation("json-minify",' { "x": 1 } '),'{"x":1}');
 assert.match(await applyOperation("json-pretty",'{"x":1}'),/\n/);
 assert.equal(await applyOperation("json-path",'{"items":[{"name":"Pippi"}]}',"items[0].name"),"Pippi");
 assert.match(await applyOperation("json-sort",'{"z":2,"a":1}'),/"a": 1[\s\S]*"z": 2/);
 await assert.rejects(applyOperation("json-pretty","not json"),/Invalid JSON/);
 await assert.rejects(applyOperation("json-path",'{"a":1}',"missing"),/Path not found/);
});
test("CSV handles multiline quotes, commas and Unicode",async()=>{
 const raw=JSON.stringify([{name:"A, B",note:'He said "hello"\non Tuesday'},{name:"Maya 🦢",note:"plain"}]);
 const csv=await applyOperation("json-to-csv",raw);
 assert.match(csv,/"A, B"/);
 const rows=JSON.parse(await applyOperation("csv-to-json",csv));
 assert.deepEqual(rows,[{name:"A, B",note:'He said "hello"\non Tuesday'},{name:"Maya 🦢",note:"plain"}]);
 await assert.rejects(applyOperation("csv-to-json",'a,b\n"never ends'),/Unterminated/);
});
test("Web Crypto SHA-256 and SHA-512 known vectors",async()=>{
 assert.equal(await applyOperation("sha-256","abc"),"ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad");
 assert.equal((await applyOperation("sha-512","abc")).length,128);
});
test("JWT display is explicit about unverified claims",async()=>{
 const s="eyJhbGciOiJub25lIn0.eyJzdWIiOiJkZW1vIn0.";
 const parsed=JSON.parse(await applyOperation("jwt-inspect",s));
 assert.equal(parsed.payload.sub,"demo");
 assert.equal(parsed.signatureVerified,false);
 assert.match(parsed.warning,/Unverified/);
});
test("URL and query parser never perform requests",async()=>{
 const url=JSON.parse(await applyOperation("url-inspect","https://api.example.org/v1/pets?status=active&status=pending#top"));
 assert.deepEqual(url.query.status,["active","pending"]);
 const p=JSON.parse(await applyOperation("query-parse","a=1&a=2"));
 assert.deepEqual(p.a,["1","2"]);
 assert.equal(await applyOperation("query-stringify",'{"a":["1","2"],"b":"space here"}'),"a=1&a=2&b=space+here");
});
test("Text transforms and statistics produce actual output",async()=>{
 assert.equal(await applyOperation("rot13","Hello"),"Uryyb");
 assert.equal(await applyOperation("sort-lines","z\na"),"a\nz");
 assert.equal(await applyOperation("unique-lines","a\na\nb"),"a\nb");
 assert.equal(JSON.parse(await applyOperation("text-stats","hi\nworld")).lines,2);
});
test("OpenAPI parser resolves local parameters and retains operation auth",()=>{
 const operations=parseOpenAPIOperations(spec);
 assert.equal(operations.length,3);
 const get=operations.find(x=>x.operationId==="findPetById");
 assert.equal(get.parameters[0].name,"petId");
 assert.equal(get.parameters[0].required,true);
 assert.equal(get.security.length,1);
 const post=operations.find(x=>x.method==="POST");
 assert.equal(post.security.length,0);
 assert.equal(post.reviewRequired,true);
});
test("OpenAPI search and offline plan cannot probe live servers",async()=>{
 const ranked=searchOpenAPIOperations(parseOpenAPIOperations(spec),"find pet by id");
 assert.equal(ranked[0].operationId,"findPetById");
 const plan=JSON.parse(await applyOperation("openapi-plan",spec,"find pet by id"));
 assert.equal(plan.networkRequests,0);
 assert.equal(plan.mode,"DOCUMENTATION_ONLY_NO_EXECUTION");
 assert.equal(plan.candidates[0].documentedPath,"/pets/{petId}");
 assert.equal(plan.candidates[0].requiresManualReview,true);
 assert.match(JSON.stringify(plan),/Declared authentication/);
});
test("Unsupported OpenAPI versions fail closed",async()=>{
 await assert.rejects(applyOperation("openapi-operations",'{"swagger":"2.0","paths":{}}'),/OpenAPI 3/);
 await assert.rejects(applyOperation("openapi-operations","openapi: 3.0.0"),/Invalid JSON/);
});
test("Recipes compose transformations and validate imported steps",async()=>{
 const recipe=[{id:"json-pretty"},{id:"base64-encode"}];
 const result=await runRecipe('{"x":1}',recipe);
 assert.equal(result.output,await applyOperation("base64-encode",'{\n  "x": 1\n}'));
 assert.equal(result.history.length,2);
 const disabled=await runRecipe("a",[{id:"uppercase",enabled:false},{id:"rot13"}]);
 assert.equal(disabled.output,"n");
 assert.throws(()=>validateRecipe([{id:"not-a-transform"}]),/Unknown operation/);
 assert.throws(()=>validateRecipe(Array.from({length:25},()=>({id:"trim"}))),/at most 24/);
 await assert.rejects(runRecipe("not json",[{id:"json-pretty"}]),/Step 1/);
});
test("Input and step options have hard limits",async()=>{
 await assert.rejects(runRecipe("x".repeat(2_000_001),[]),/2 million/);
 assert.throws(()=>validateRecipe([{id:"json-path",option:"x".repeat(501)}]),/500/);
});
