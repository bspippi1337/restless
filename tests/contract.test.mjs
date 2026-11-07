import test from "node:test";
import assert from "node:assert/strict";
import {inspectContract,compareContracts,findOperations} from "../site/contract/engine.mjs";

const spec=(version,paths,extra={})=>JSON.stringify({
  openapi:"3.0.3",info:{title:"Pets API",version},servers:[{url:"https://example.invalid"}],
  paths,...extra
});
const old=spec("1.0.0",{
  "/pets":{get:{operationId:"listPets",summary:"List pets",tags:["pets"]}},
  "/pets/{petId}":{get:{operationId:"getPet",summary:"Get pet",parameters:[{name:"petId",in:"path",required:true}]}}
});
test("Catalog inspections have no network calls and count HTTP operations",()=>{
  const report=inspectContract(old);
  assert.equal(report.networkRequests,0);
  assert.equal(report.stats.operations,2);
  assert.equal(report.stats.uniquePaths,2);
  assert.equal(report.stats.methods.GET,2);
  assert.equal(report.api.title,"Pets API");
  assert.equal(report.findings.filter(f=>f.severity==="review").length,0);
});
test("Reviews missing path parameters and unresolved external references",()=>{
  const source=spec("2.0.0",{
    "/pets/{petId}":{patch:{summary:"Update pet",parameters:[{$ref:"https://evil.invalid/param"}]}}
  });
  const report=inspectContract(source);
  assert.ok(report.findings.some(x=>x.code==="UNDECLARED_PATH_PARAMETER"));
  assert.ok(report.findings.some(x=>x.code==="UNRESOLVED_REF"));
  assert.equal(report.stats.needsReview,1);
  assert.equal(report.networkRequests,0);
});
test("Diff classifies removal as potentially breaking and additions as review",()=>{
  const current=spec("1.1.0",{
    "/pets":{get:{operationId:"listPets",summary:"List pets",tags:["pets"]}},
    "/orders":{post:{operationId:"makeOrder",summary:"Create order"}}
  });
  const diff=compareContracts(old,current);
  assert.equal(diff.networkRequests,0);
  assert.equal(diff.counts.added,1);
  assert.equal(diff.counts.removed,1);
  assert.equal(diff.counts.unchanged,1);
  assert.equal(diff.counts.potentiallyBreaking,1);
  assert.equal(diff.changes.find(c=>c.type==="removed").risk,"potentially-breaking");
});
test("Newly required parameters produce a potential breaking signal",()=>{
  const before=spec("1", {"/items":{get:{operationId:"list",summary:"Items"}}});
  const after=spec("2", {"/items":{get:{operationId:"list",summary:"Items",parameters:[{in:"query",name:"account",required:true}]}}});
  const report=compareContracts(before,after);
  assert.equal(report.counts.changed,1);
  assert.equal(report.counts.potentiallyBreaking,1);
  assert.ok(report.changes[0].reasons.some(x=>/Newly required/.test(x)));
});
test("A change in security declarations is reported but no auth is attempted",()=>{
  const before=spec("1", {"/items":{get:{operationId:"list",summary:"Items"}}});
  const after=spec("2", {"/items":{get:{operationId:"list",summary:"Items",security:[{api_key:[]}]}}});
  const diff=compareContracts(before,after);
  assert.equal(diff.counts.potentiallyBreaking,1);
  assert.equal(diff.changes[0].risk,"potentially-breaking");
  assert.equal(diff.networkRequests,0);
});
test("Reordering tags does not produce a false contract change",()=>{
  const before=spec("1",{"/items":{get:{summary:"Items",tags:["catalog","read"]}}});
  const after=spec("2",{"/items":{get:{summary:"Items",tags:["read","catalog"]}}});
  const diff=compareContracts(before,after);
  assert.equal(diff.counts.unchanged,1);
  assert.equal(diff.counts.changed,0);
});
test("Find operations supports lexical search and method filters",()=>{
  const report=inspectContract(spec("1",{"/orders":{post:{operationId:"createOrder",summary:"Create order"}}, "/orders/{id}":{get:{operationId:"findOrder",summary:"Find order by ID"}}}));
  const all=findOperations(report,"order");
  assert.equal(all.length,2);
  const get=findOperations(report,"order","GET");
  assert.equal(get.length,1);
  const review=findOperations(report,"order","ALL",true);
  assert.ok(review.some(op=>op.method==="POST"));
});
test("Invalid input, missing paths and other OpenAPI versions fail safely",()=>{
  assert.throws(()=>inspectContract("not json"),/Invalid OpenAPI JSON/);
  assert.throws(()=>inspectContract('{"swagger":"2.0","paths":{}}'),/OpenAPI 3/);
  assert.throws(()=>inspectContract('{"openapi":"3.0.3"}'),/paths/);
  assert.throws(()=>inspectContract("x".repeat(2_000_001)),/2 million/);
});
