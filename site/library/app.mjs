import {validateDomain,readAPI,makeTree,makeExample,buildRequestPreview,toPostman,toMermaid} from "./engine.mjs";
const $=id=>document.getElementById(id),NS="http://www.w3.org/2000/svg";
const state={model:null,tree:null,selected:null,search:"",method:"ALL",domain:null,lastCommand:"",tab:"parameters",inputs:Object.create(null),bodyText:"",code:"curl"};
function el(tag,cls,text){const n=document.createElement(tag);if(cls)n.className=cls;if(text!==undefined)n.textContent=String(text);return n;}
function svg(tag,attrs){const n=document.createElementNS(NS,tag);for(const[k,v]of Object.entries(attrs||{}))n.setAttribute(k,String(v));return n;}
function notice(message,kind="info"){$("atlas-status").textContent=message;$("atlas-status").dataset.kind=kind;}
function download(name,contents,mime="application/json;charset=utf-8"){
 const href=URL.createObjectURL(new Blob([contents],{type:mime})),a=el("a");
 a.href=href;a.download=name;document.body.append(a);a.click();a.remove();setTimeout(()=>URL.revokeObjectURL(href),5000);
}
async function copy(text){
 try{await navigator.clipboard.writeText(text);notice("Copied to clipboard.","success");}
 catch{notice("Clipboard unavailable. Use download or select the text manually.","error");}
}
const petBody={content:{"application/json":{schema:{
 type:"object",properties:{name:{type:"string",example:"Pixel"},status:{type:"string",enum:["available","adopted"]}}
}}}};
const demo={
 openapi:"3.0.3",
 info:{title:"RESTLESS Demo API",version:"2.4.0",description:"Synthetic demo, not an actual server"},
 servers:[{url:"https://demo.example.invalid/v2"}],
 paths:{
  "/pets":{
   get:{operationId:"listPets",tags:["Pets"],summary:"List pets",parameters:[{name:"status",in:"query",schema:{type:"string",enum:["available","pending"]}}]},
   post:{operationId:"createPet",tags:["Pets"],summary:"Register a pet",requestBody:petBody}
  },
  "/pets/{petId}":{
   parameters:[{name:"petId",in:"path",required:true,schema:{type:"integer",example:42}}],
   get:{operationId:"findPet",tags:["Pets"],summary:"Find pet by ID"},
   patch:{operationId:"updatePet",tags:["Pets"],summary:"Change pet status",requestBody:petBody},
   delete:{operationId:"removePet",tags:["Pets"],summary:"Remove a pet"}
  },
  "/orders":{
   get:{operationId:"listOrders",tags:["Orders"],summary:"List orders"},
   post:{operationId:"createOrder",tags:["Orders"],summary:"Create an order",requestBody:petBody}
  },
  "/accounts/{accountId}/orders":{
   parameters:[{name:"accountId",in:"path",required:true,schema:{type:"string"}}],
   get:{operationId:"accountOrders",tags:["Accounts"],summary:"Find account orders"}
  },
  "/health":{head:{operationId:"checkHealth",tags:["System"],summary:"Service health headers"}}
 }
};
function inspectDomain(){
 try{
  const d=validateDomain($("domain").value);state.domain=d;
  state.lastCommand="restless library "+d.url+" --max-pages 12 --max-depth 2 --out .restless/libraries --execute --authorized";
  $("domain-output").hidden=false;$("domain-command").textContent=state.lastCommand;
  $("domain-badge").textContent="EXACT HOST / "+d.host;
  $("domain-note").textContent=d.explicitPath?
   "The native crawler follows linked documentation. Browser fetch can read this one explicit JSON document only when the host allows CORS.":
   "For domain-root discovery use native RESTLESS. Browser fetch is disabled until a specific documentation URL is supplied.";
  $("authorized").checked=false;$("fetch-doc").disabled=true;
  notice("Scope preview complete. ZERO network requests. Nothing was guessed.","success");
 }catch(e){state.domain=null;$("domain-output").hidden=true;notice(e.message,"error");}
}
function allowFetch(){$("fetch-doc").disabled=!state.domain||!state.domain.explicitPath||!$("authorized").checked;}
async function fetchExplicit(){
 if(!state.domain||!state.domain.explicitPath||!$("authorized").checked){notice("Authorization and an exact HTTPS document URL are required.","error");return;}
 const d=state.domain;
 $("fetch-doc").disabled=true;notice("Requesting one explicit JSON file from "+d.host+" under browser CORS rules.");
 try{
  const controller=new AbortController(),timer=setTimeout(()=>controller.abort(),12000);let response;
  try{response=await fetch(d.url,{method:"GET",redirect:"error",mode:"cors",credentials:"omit",cache:"no-store",referrerPolicy:"no-referrer",signal:controller.signal,headers:{Accept:"application/json"}});}
  finally{clearTimeout(timer);}
  if(!response.ok)throw Error("Documentation returned HTTP "+response.status);
  if(Number(response.headers.get("content-length"))>2_000_000)throw Error("Document exceeds 2 MB.");
  const reader=response.body?.getReader();if(!reader)throw Error("No readable response body.");
  let bytes=0,chunks=[];
  try{
   while(true){const{done,value}=await reader.read();if(done)break;bytes+=value.byteLength;if(bytes>2_000_000)throw Error("Document exceeds 2 MB.");chunks.push(value);}
  }catch(e){try{await reader.cancel();}catch{}throw e;}finally{reader.releaseLock();}
  const data=new Uint8Array(bytes);let offset=0;for(const part of chunks){data.set(part,offset);offset+=part.length;}
  const model=readAPI(new TextDecoder("utf-8",{fatal:true}).decode(data));
  if(model.kind!=="OPENAPI_3_JSON")throw Error("The browser fetch supports only OpenAPI 3 JSON.");
  load(model,"Imported one explicit HTTPS document, no API methods executed.");
 }catch(e){notice("Documentation import failed: "+e.message+". Use the native crawler and import library.json if CORS blocks the browser.","error");}
 finally{allowFetch();}
}
async function importFile(file){
 if(!file)return;if(file.size>2_000_000){notice("File exceeds 2 MB.","error");return;}
 try{load(readAPI(await file.text()),"Imported "+file.name+" locally. No upload.");}
 catch(e){notice("Import failed: "+e.message,"error");}
}
function load(model,source){
 state.model=model;state.tree=makeTree(model);state.selected=model.operations[0]||null;
 state.search="";state.method="ALL";$("search").value="";
 $("dashboard").hidden=false;$("api-title").textContent=model.apiName||"API Library";
 $("api-meta").textContent=model.kind.replaceAll("_"," ")+" · "+model.version+" · "+source;
 $("stat-bar").replaceChildren();
 for(const[label,value]of [["Operations",model.stats.operations],["Documented paths",model.stats.paths],["Evidence documents",model.stats.documents],["HTTP methods",Object.keys(model.stats.methods).length]]){
  const card=el("div","stat-card");card.append(el("strong","",value),el("span","",label));$("stat-bar").append(card);
 }
 renderFilters();renderResources();renderGraph();renderInspector();
 notice("Mapped "+model.stats.operations+" documented operations across "+state.tree.nodes.length+" resource nodes. Zero API operations executed.","success");
 $("dashboard").scrollIntoView({behavior:"smooth",block:"start"});
}
function match(op){
 const query=state.search.toLowerCase();
 return(state.method==="ALL"||op.method===state.method)&&(!query||[op.path,op.operationId,op.summary,op.description,...op.tags].join(" ").toLowerCase().includes(query));
}
function renderFilters(){
 const strip=$("method-strip");strip.replaceChildren();
 for(const method of ["ALL",...Object.keys(state.model.stats.methods).sort()]){
  const b=el("button","method-filter",method==="ALL"?"ALL":method+" "+state.model.stats.methods[method]);
  b.type="button";b.setAttribute("aria-pressed",String(method===state.method));
  b.addEventListener("click",()=>{state.method=method;renderFilters();renderResources();});strip.append(b);
 }
}
function renderResources(){
 const list=$("resource-list");list.replaceChildren();const groups=new Map();
 for(const op of state.model.operations.filter(match)){if(!groups.has(op.path))groups.set(op.path,[]);groups.get(op.path).push(op);}
 $("resource-total").textContent=groups.size+" paths";
 for(const[path,ops]of groups){
  const group=el("div","resource-group"),head=el("div","resource-name");
  head.append(el("span","",path),el("span","",ops.length));group.append(head);
  for(const op of ops){
   const b=el("button","operation-button");b.type="button";b.setAttribute("aria-selected",String(op===state.selected));
   b.setAttribute("aria-label","Inspect "+op.method+" "+op.path);
   b.append(el("span","method-chip method-"+op.method.toLowerCase(),op.method),
            el("span","summary",op.summary||op.operationId||op.path));
   b.addEventListener("click",()=>select(op));group.append(b);
  }list.append(group);
 }
 if(!groups.size)list.append(el("p","input-hint","No matching documented operations."));
}
function select(op){state.selected=op;state.inputs=Object.create(null);state.bodyText="";renderResources();renderGraph();renderInspector();}
function renderGraph(){
 const tree=state.tree,canvas=$("api-graph");canvas.replaceChildren();
 if(!tree.nodes.length)return;
 const positions=new Map(),ordered=[];let leaf=0;
 function layout(n){
  for(const child of n.children.sort((a,b)=>a.name.localeCompare(b.name)))layout(child);
  const y=n.children.length?(positions.get(n.children[0].id).y+positions.get(n.children[n.children.length-1].id).y)/2:85+leaf++*79;
  positions.set(n.id,{x:30+n.depth*217,y});ordered.push(n);
 }
 layout(tree.root);
 const width=Math.max(760,tree.maxDepth*217+255),height=Math.max(450,Math.min(30000,leaf*79+110));
 canvas.setAttribute("viewBox","0 0 "+width+" "+height);canvas.setAttribute("width",width);canvas.setAttribute("height",height);
 for(const edge of tree.edges){
  const a=positions.get(edge.from),b=positions.get(edge.to);if(!a||!b)continue;
  canvas.append(svg("path",{class:"graph-edge",d:"M "+(a.x+172)+" "+a.y+" C "+(a.x+191)+" "+a.y+" "+(b.x-20)+" "+b.y+" "+b.x+" "+b.y}));
 }
 for(const n of ordered){
  const p=positions.get(n.id),chosen=state.selected?.path===n.id;
  const g=svg("g",{class:"graph-node"+(chosen?" selected":""),tabindex:"0",role:"button","aria-label":"Inspect resource "+n.path});
  g.append(svg("rect",{x:p.x,y:p.y-24,width:172,height:50}));
  const title=svg("text",{class:"graph-label",x:p.x+10,y:p.y-3});title.textContent=n.name.length>19?n.name.slice(0,17)+"…":n.name;
  const sub=svg("text",{class:"graph-sub",x:p.x+10,y:p.y+13});
  sub.textContent=n.operations.length?n.operations.map(o=>o.method).join(" · ").slice(0,25):n.children.length+" documented branches";
  g.append(title,sub);
  if(n.operations.length)g.append(svg("circle",{class:"graph-link-dot",cx:p.x+161,cy:p.y-15,r:3.4}));
  const activate=()=>{const candidate=n.operations[0]||state.model.operations.find(o=>o.path.startsWith(n.path==="/"?"/":n.path+"/"));if(candidate)select(candidate);};
  g.addEventListener("click",activate);
  g.addEventListener("keydown",e=>{if(e.key==="Enter"||e.key===" "){e.preventDefault();activate();}});
  canvas.append(g);
 }
}
function field(parent,label,value,asCode=false){
 const s=el("div","source-flag");s.append(el("strong","",label),el(asCode?"code":"span","",value));parent.append(s);
}
function renderInspector(){
 const op=state.selected;$("inspector-empty").hidden=!!op;$("inspector-content").hidden=!op;
 if(!op){$("operation-indicator").textContent="SELECT";return;}
 $("operation-indicator").textContent="DOCUMENTED";
 $("op-method").textContent=op.method;$("op-method").className="method-chip method-"+op.method.toLowerCase();
 $("op-path").textContent=op.path;$("op-name").textContent=op.summary||op.operationId||op.method+" "+op.path;
 $("op-description").textContent=op.description||"Declared in the imported source. No live endpoint verification.";
 const example=makeExample(state.model,op);
 const inputs=$("tab-parameters");inputs.replaceChildren();
 if(!example.parameters.length)inputs.append(el("p","input-hint","No parameters declared in this operation."));
 for(const p of example.parameters){
   const box=el("div","param-row"),header=el("div","param-title");
   header.append(el("strong","",p.name||"unnamed"),el("span","",p.in+" / "+p.type+(p.required?" / required":"")));
   const label=el("label","input-label","Enter "+p.in+" value for "+p.name);
   const fieldInput=el("input","example-input");
   fieldInput.type="text";fieldInput.placeholder=String(p.example);fieldInput.value=state.inputs[p.in+":"+p.name]||"";
   fieldInput.maxLength=500;fieldInput.setAttribute("aria-label",label.textContent);
   fieldInput.addEventListener("input",()=>{state.inputs[p.in+":"+p.name]=fieldInput.value;updateCode();});
   box.append(header,el("p","",p.description||"Documented parameter."),label,fieldInput);
   inputs.append(box);
 }
 if(example.body!==null){
   const box=el("div","param-row");
   const label=el("label","input-label","Example request body JSON");
   const textarea=el("textarea","body-input");
   textarea.value=state.bodyText||JSON.stringify(example.body,null,2);
   textarea.setAttribute("aria-label","Editable JSON request body");
   textarea.addEventListener("input",()=>{state.bodyText=textarea.value;updateCode();});
   box.append(el("strong","",example.contentType||"Documented JSON request body"),label,textarea);inputs.append(box);
 }
 const evidence=$("tab-source");evidence.replaceChildren();
 field(evidence,"SOURCE DOCUMENT",op.sourceURL||"Local OpenAPI 3 JSON import",true);
 field(evidence,"SHA-256 PROVENANCE",op.sourceSHA256||"Unavailable for pasted specification",true);
 field(evidence,"DECLARED AUTH",JSON.stringify(op.security)+" (not authenticated)",true);
 field(evidence,"UNRESOLVED REFERENCES",String(op.unresolvedRefs));
 field(evidence,"REVIEW REQUIRED",String(example.reviewRequired));
 field(evidence,"ACTUAL API EXECUTIONS","ZERO. Documentation-derived only.");
 updateCode();renderTab();
}
function updateCode(){
 if(!state.model||!state.selected)return;
 try{
   const preview=buildRequestPreview(state.model,state.selected,{values:state.inputs,bodyText:state.bodyText});
   const code=preview[state.code]||preview.curl;
   $("curl-example").textContent=code;
   $("code-notice").textContent=preview.missing.length?
     "Unfilled required inputs: "+preview.missing.join(", ")+". Code is a preview only.":
     "Preview constructed locally. No HTTP request performed; review permission, host and possible effects.";
   $("copy-curl").disabled=false;
 }catch(e){
   $("curl-example").textContent="";
   $("code-notice").textContent="Input validation: "+e.message;
   $("copy-curl").disabled=true;
 }
}
function renderTab(){document.querySelectorAll(".inspector-tabs button").forEach(t=>{const active=t.dataset.tab===state.tab;t.setAttribute("aria-selected",String(active));$("tab-"+t.dataset.tab).hidden=!active;});}
function basename(){return(state.model.apiName||"api").replace(/[^a-zA-Z0-9_-]+/g,"-").slice(0,60).toLowerCase();}
function saveGraph(){
 const graph=$("api-graph");if(!state.model||!graph.children.length)return;
 const clone=graph.cloneNode(true),style=svg("style");
 style.textContent=".graph-edge{fill:none;stroke:#7aa68f;stroke-width:2}.graph-node rect{fill:#1d3040;stroke:#63859b;rx:9}.graph-node.selected rect{fill:#28513b;stroke:#b7f768;stroke-width:2}.graph-label{fill:#eef8ed;font:700 12px monospace}.graph-sub{fill:#a2bdca;font:10px monospace}.graph-link-dot{fill:#b7f768}";
 clone.insertBefore(style,clone.firstChild);
 clone.insertBefore(svg("rect",{x:0,y:0,width:"100%",height:"100%",fill:"#0b1722"}),clone.firstChild);
 download(basename()+"-map.svg",new XMLSerializer().serializeToString(clone),"image/svg+xml;charset=utf-8");
}
$("preview").addEventListener("click",inspectDomain);
$("domain").addEventListener("keydown",e=>{if(e.key==="Enter")inspectDomain();});
$("authorized").addEventListener("change",allowFetch);
$("fetch-doc").addEventListener("click",fetchExplicit);
$("copy-command").addEventListener("click",()=>copy(state.lastCommand));
$("load-demo").addEventListener("click",()=>load(readAPI(demo),"Synthetic local demonstration. No servers contacted."));
$("open-json").addEventListener("click",()=>$("json-file").click());
$("json-file").addEventListener("change",async e=>{const file=e.target.files?.[0];e.target.value="";await importFile(file);});
$("search").addEventListener("input",()=>{state.search=$("search").value.trim();renderResources();});
$("reset-graph").addEventListener("click",()=>$("graph-viewport").scrollTo({top:0,left:0,behavior:"smooth"}));
$("save-graph").addEventListener("click",saveGraph);
$("export-postman").addEventListener("click",()=>download(basename()+"-postman.json",JSON.stringify(toPostman(state.model),null,2)));
$("export-mermaid").addEventListener("click",()=>download(basename()+"-map.mmd",toMermaid(state.model),"text/plain;charset=utf-8"));
$("export-json").addEventListener("click",()=>download(basename()+"-atlas.json",JSON.stringify(state.model,null,2)));
$("copy-curl").addEventListener("click",()=>{if(state.selected)copy($("curl-example").textContent);});
$("code-language").addEventListener("change",()=>{state.code=$("code-language").value;updateCode();});
document.querySelectorAll(".inspector-tabs button").forEach(b=>b.addEventListener("click",()=>{state.tab=b.dataset.tab;renderTab();}));
notice("No source loaded. This browser never guesses API paths. Import a real library.json or OpenAPI 3 JSON.");
