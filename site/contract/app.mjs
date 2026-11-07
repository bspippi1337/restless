import {inspectContract,compareContracts,findOperations} from "./engine.mjs";

const $=id=>document.getElementById(id);
const state={report:null,mode:"explore"};
const baseline=$("baseline"),candidate=$("candidate");
const status=$("contract-status"),details=$("details-json"),detailPanel=document.querySelector(".json-details");
const result=$("results"),summary=$("summary");
function node(tag,className,text){
 const el=document.createElement(tag);
 if(className)el.className=className;
 if(text!==undefined)el.textContent=String(text);
 return el;
}
function setStatus(message,kind="info"){
 status.textContent=message;status.dataset.kind=kind;
}
function summarize(label,value){
 const card=node("div","summary-stat");
 card.append(node("strong","",value),node("small","",label));
 return card;
}
function setInfo(report,mode){
 state.report=report;state.mode=mode;
 result.hidden=false;$("export-report").disabled=false;
 details.textContent="Select a row to inspect the documented evidence.";
 detailPanel.open=false;
 $("explore-section").hidden=mode!=="explore";
 $("compare-section").hidden=mode!=="compare";
 $("result-eyebrow").textContent=mode==="explore"?"DOCUMENTATION INSPECTION":"OFFLINE CONTRACT COMPARISON";
 $("result-heading").textContent=mode==="explore"?report.api.title:"What changed in the contract";
 $("result-subtitle").textContent=mode==="explore"?report.api.version+" · OpenAPI "+report.api.openapi:report.baseline.version+" → "+report.candidate.version;
 summary.replaceChildren();
 if(mode==="explore"){
   const s=report.stats;
   summary.append(summarize("Documented operations",s.operations),
     summarize("Distinct paths",s.uniquePaths),
     summarize("Declared security",s.securityDeclared),
     summarize("Review flags",s.needsReview));
 }else{
   const c=report.counts;
   summary.append(summarize("Added",c.added),summarize("Removed",c.removed),
     summarize("Changed",c.changed),summarize("Potential breaking risk",c.potentiallyBreaking));
 }
}
function cell(text,className){
 const td=node("td",className||"",text);
 return td;
}
function selected(record){
 details.textContent=JSON.stringify(record,null,2);
 detailPanel.open=true;
 detailPanel.scrollIntoView({behavior:"smooth",block:"nearest"});
}
function operationRow(op){
 const tr=node("tr");
 const method=cell();method.append(node("span","method method-"+op.method.toLowerCase(),op.method));
 const path=cell(op.path,"path");
 const id=cell(op.operationId||op.summary||"(not named)");
 const risk=cell(op.reviewRequired?"Manual review":"Review as needed","risk "+(!op.reviewRequired?"safe":""));
 const control=cell();const b=node("button","select-record","Inspect");
 b.type="button";b.setAttribute("aria-label","Inspect "+op.method+" "+op.path);
 b.addEventListener("click",()=>selected(op));control.append(b);
 tr.append(method,path,id,risk,control);
 return tr;
}
function updateFilter(){
 if(state.mode!=="explore"||!state.report)return;
 const query=$("operation-filter").value,method=$("method-filter").value,review=$("review-only").checked;
 const filtered=findOperations(state.report,query,method,review);
 $("operation-total").textContent=filtered.length+" shown";
 const body=$("operations-body");body.replaceChildren();
 for(const op of filtered)body.append(operationRow(op));
 if(!filtered.length){const row=node("tr");const empty=cell("No documented operations matched these filters.");empty.colSpan=5;row.append(empty);body.append(row);}
}
function showFindings(findings){
 const wrap=$("findings");wrap.replaceChildren();
 if(!findings.length){wrap.append(node("p","","No specific documentation warnings were detected by this basic static audit. This is not a guarantee of correctness."));return;}
 for(const f of findings){
   const box=node("div","finding"+(f.severity==="review"?" review":""));
   box.append(node("strong","",f.code+(f.operation?" · "+f.operation:"")),node("p","",f.message));
   wrap.append(box);
 }
}
function changeRow(change){
 const tr=node("tr");
 const label=cell(change.type.toUpperCase());
 const op=cell(change.key,"path");
 const risk=cell(change.risk,"risk "+(change.risk==="none"?"safe":""));
 const explanation=cell(change.reasons.join("; "));
 const controls=cell();const b=node("button","select-record","Inspect");
 b.type="button";b.setAttribute("aria-label","Inspect "+change.key+" change");
 b.addEventListener("click",()=>selected(change));controls.append(b);
 tr.append(label,op,risk,explanation,controls);
 return tr;
}
function renderDiff(report){
 const body=$("changes-body");body.replaceChildren();
 for(const change of report.changes)body.append(changeRow(change));
 if(!report.changes.length){const tr=node("tr");const td=cell("No changed operations between these documented specifications.");td.colSpan=5;tr.append(td);body.append(tr);}
 $("change-total").textContent=report.changes.length+" differences · "+report.counts.unchanged+" unchanged";
}
function run(mode){
 try{
   if(!candidate.value.trim())throw Error("Paste or import a candidate OpenAPI 3 JSON specification.");
   const start=performance.now();
   let report;
   if(mode==="compare"){
     if(!baseline.value.trim())throw Error("Provide a baseline OpenAPI JSON document to compare.");
     report=compareContracts(baseline.value,candidate.value);
   }else report=inspectContract(candidate.value);
   setInfo(report,mode);
   if(mode==="explore"){updateFilter();showFindings(report.findings);}
   else renderDiff(report);
   const ms=Math.round(performance.now()-start);
   setStatus((mode==="explore"?"Catalog inspected.":"Contracts compared.")+" "+ms+" ms. All processing stayed in this browser; zero network requests.","success");
   result.scrollIntoView({behavior:"smooth",block:"start"});
 }catch(e){
   result.hidden=true;state.report=null;$("export-report").disabled=true;
   setStatus(e.message||String(e),"error");
 }
}
function download(name,data){
 const url=URL.createObjectURL(new Blob([data],{type:"application/json;charset=utf-8"}));
 const a=node("a");a.href=url;a.download=name;document.body.append(a);a.click();a.remove();
 setTimeout(()=>URL.revokeObjectURL(url),5000);
}
function preview(){
 const previous={openapi:"3.0.3",info:{title:"RESTLESS Example API",version:"1.0.0"},servers:[{url:"https://example.invalid"}],
  paths:{
   "/pets":{get:{operationId:"listPets",summary:"List pets",tags:["pets"]}},
   "/pets/{petId}":{get:{operationId:"getPet",summary:"Find pet by ID",tags:["pets"],parameters:[{in:"path",name:"petId",required:true}]}}
  }};
 const current={openapi:"3.0.3",info:{title:"RESTLESS Example API",version:"2.0.0"},servers:[{url:"https://example.invalid"}],
  security:[{api_key:[]}],
  paths:{
   "/pets":{get:{operationId:"listPets",summary:"List all pets including archived",tags:["pets"],parameters:[{in:"query",name:"tenant",required:true}]}},
   "/orders":{post:{operationId:"createOrder",summary:"Create a customer order",tags:["orders"],security:[{oauth2:["write:orders"]}]}},
   "/pets/{petId}":{get:{operationId:"getPet",summary:"Find pet by ID",tags:["pets"],parameters:[{in:"path",name:"petId",required:true}]}}
  }};
 baseline.value=JSON.stringify(previous,null,2);candidate.value=JSON.stringify(current,null,2);
 run("compare");
}
function loadFile(input,box){
 input.addEventListener("change",async()=>{
  const file=input.files?.[0];input.value="";if(!file)return;
  if(file.size>2_000_000){setStatus("File exceeds 2 MB; nothing imported.","error");return;}
  try{
   const text=await file.text();
   if(text.length>2_000_000)throw Error("Document exceeds 2 million characters.");
   box.value=text;setStatus("Loaded "+file.name+" into a local text field. No upload.", "success");
  }catch(e){setStatus("Cannot import document: "+e.message,"error");}
 });
}
$("import-before").addEventListener("click",()=>$("file-before").click());
$("import-after").addEventListener("click",()=>$("file-after").click());
loadFile($("file-before"),baseline);loadFile($("file-after"),candidate);
$("example").addEventListener("click",preview);
$("clear").addEventListener("click",()=>{
 baseline.value="";candidate.value="";
 result.hidden=true;state.report=null;$("export-report").disabled=true;
 setStatus("Documents cleared. No source text is persisted.","success");
});
$("analyze").addEventListener("click",()=>run("explore"));
$("compare").addEventListener("click",()=>run("compare"));
$("export-report").addEventListener("click",()=>{
 if(!state.report)return;
 download(state.mode==="explore"?"restless42-openapi-inspection.json":"restless42-contract-diff.json",JSON.stringify(state.report,null,2));
 setStatus("JSON analysis report downloaded. Raw source documents are not embedded automatically.","success");
});
$("operation-filter").addEventListener("input",updateFilter);
$("method-filter").addEventListener("change",updateFilter);
$("review-only").addEventListener("change",updateFilter);
setStatus("Ready. Import OpenAPI 3 JSON documents, or load demo revisions.");
