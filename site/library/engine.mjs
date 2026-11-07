/**
 * RESTLESS 42 API Atlas: local parser, methodical API tree, request examples,
 * Postman Collection 2.1 export. Never performs network traffic.
 * Both Go library.json and OpenAPI 3 JSON are accepted.
 */
import {parseOpenAPIOperations} from "../toolbox/transforms.mjs";
const MAX_TEXT=2_000_000,MAX_OPERATIONS=1000;
const METHODS=["GET","HEAD","OPTIONS","POST","PUT","PATCH","DELETE","TRACE"];
const SAFE_HOST=/^[a-z0-9](?:[a-z0-9.-]{0,251}[a-z0-9])?$/;
const RESERVED=[".local",".localhost",".internal",".home",".lan",".test",".example",".invalid",".onion"];
function isObj(x){return x!==null&&typeof x==="object"&&!Array.isArray(x);}
export function validateDomain(text){
 if(typeof text!=="string"||!text.trim()||text.length>2048)throw Error("Enter a domain or an explicit HTTPS documentation URL.");
 const value=text.trim();
 if(/\s/.test(value))throw Error("Domain contains whitespace.");
 const authority=value.replace(/^https:\/\//i,"").split(/[/?#]/)[0];
 if(/:[0-9]+$/.test(authority))throw Error("Explicit ports are not supported.");
 let u;try{u=new URL(value.includes("://")?value:"https://"+value);}catch{throw Error("Invalid domain or HTTPS documentation URL.");}
 if(u.protocol!=="https:")throw Error("HTTPS is required.");
 if(u.username||u.password||u.port||u.search||u.hash)throw Error("Credentials, nonstandard ports, queries and fragments are not supported.");
 const host=u.hostname.toLowerCase();
 if(!SAFE_HOST.test(host)||!host.includes(".")||host.includes("..")||host.endsWith(".")||RESERVED.some(x=>host.endsWith(x))||/^\d+(?:\.\d+){3}$/.test(host)||host.includes(":"))throw Error("Enter a public DNS hostname, not an IP or internal network.");
 return {host,url:u.href,explicitPath:u.pathname!=="/"};
}
function parseJSON(text){
 if(typeof text!=="string"||text.length>MAX_TEXT)throw Error("Input must be JSON under 2 million characters.");
 let x;try{x=JSON.parse(text);}catch(e){throw Error("Invalid JSON: "+e.message);}
 if(!isObj(x))throw Error("Expected a JSON document.");
 return x;
}
function normalParam(p){
 return {name:String(p?.name||""),in:String(p?.in||""),required:!!p?.required,description:String(p?.description||"").slice(0,350),
 type:String(p?.type||p?.schema?.type||"string"),schema:isObj(p?.schema)?p.schema:{},unresolvedRef:!!p?.unresolvedRef};
}
function normalizeOperation(raw,source,sha){
 if(!METHODS.includes(String(raw?.method||"").toUpperCase())||typeof raw?.path!=="string"||!raw.path.startsWith("/"))return null;
 if(raw.path.length>1024||raw.path.split("/").length>34)throw Error("Documented API path exceeds the 1,024-character or 32-segment safety limit.");
 return {
   method:raw.method.toUpperCase(),path:raw.path,operationId:String(raw.operationId||raw.operation_id||""),
   summary:String(raw.summary||""),description:String(raw.description||"").slice(0,2000),
   tags:Array.isArray(raw.tags)?raw.tags.filter(x=>typeof x==="string").slice(0,30):[],
   parameters:Array.isArray(raw.parameters)?raw.parameters.slice(0,200).map(normalParam):[],
   security:Array.isArray(raw.security)?raw.security:[],authSpecified:!!raw.auth_specified,
   sourceURL:String(raw.source_url||source||""),sourceSHA256:String(raw.source_sha256||sha||""),
   unresolvedRefs:!!(raw.unresolved_refs||raw.unresolvedRefs),requestBody:isObj(raw.requestBody)?raw.requestBody:null,
   request:isObj(raw.request)?raw.request:null,
   responses:Array.isArray(raw.responses)?raw.responses:isObj(raw.responses)?raw.responses:null
 };
}
function documentedServer(spec){
 const servers=Array.isArray(spec.servers)?spec.servers:[];
 const chosen=servers.find(s=>typeof s?.url==="string"&&s.url.startsWith("https://")&&!s.url.includes("{"));
 if(!chosen)return "";
 try{const u=new URL(chosen.url);return u.protocol==="https:"&&!u.username&&!u.password?u.origin+u.pathname.replace(/\/$/,""):"";}
 catch{return "";}
}
export function readAPI(input){
 const doc=typeof input==="string"?parseJSON(input):input;
 if(!isObj(doc))throw Error("Expected a JSON object.");
 let operations=[],documents=[],kind="",apiName="",version="",host="",server="",warnings=[];
 if(typeof doc.openapi==="string"&&doc.openapi.startsWith("3.")){
   kind="OPENAPI_3_JSON";apiName=String(doc.info?.title||"Untitled API");version=String(doc.info?.version||"not specified");
   server=documentedServer(doc);
   const rawOps=parseOpenAPIOperations(doc);
   for(const op of rawOps){
     const raw=doc.paths?.[op.path]?.[op.method.toLowerCase()]||{};
     const params=op.parameters.map(p=>{
       const declared=[...(doc.paths?.[op.path]?.parameters||[]),...(raw.parameters||[])].find(x=>x?.name===p.name&&x?.in===p.in);
       return {...p,schema:declared?.schema||{},type:declared?.schema?.type||"string"};
     });
     const bodyHasUnresolvedSchema=isObj(raw.requestBody?.content)&&
       Object.values(raw.requestBody.content).some(item=>typeof item?.schema?.$ref==="string");
     const o=normalizeOperation({...op,parameters:params,unresolvedRefs:op.unresolvedRefs||bodyHasUnresolvedSchema,
       requestBody:raw.requestBody,responses:raw.responses},"","");
     if(o)operations.push(o);
   }
   documents=[{title:apiName,kind:"openapi-3",url:"Local JSON import",sha256:""}];
   warnings.push("Source is a user-supplied OpenAPI JSON file, not a live deployment probe.");
 }else if(doc.schema_version===1&&isObj(doc.target)&&Array.isArray(doc.operations)){
   kind="RESTLESS_LIBRARY";apiName=String(doc.target.host||"API library");host=String(doc.target.host||"");
   server=host?"https://"+host:"";
   version=String(doc.generated_at||"");
   operations=doc.operations.map(x=>normalizeOperation(x,x?.source_url,x?.source_sha256)).filter(Boolean);
   documents=Array.isArray(doc.documents)?doc.documents.filter(isObj).slice(0,1000):[];
   warnings.push("Imported from RESTLESS crawler export: SHA-256 values are supplied by the archive, but the browser cannot verify them without original archived bytes.");
   warnings.push("The library records documented methods. It has not called the documented operations.");
 }else if(doc.version===1&&Array.isArray(doc.operations)&&Array.isArray(doc.documents)){
   kind="RESTLESS_CATALOG";apiName="Imported API catalog";
   operations=doc.operations.map(x=>normalizeOperation(x,x?.source_url,x?.source_sha256)).filter(Boolean);
   documents=doc.documents.filter(isObj).slice(0,1000);
   warnings.push("Catalog imported locally. Method labels are documentation, not executed calls.");
 }else throw Error("Expected OpenAPI 3 JSON, RESTLESS library.json, or RESTLESS catalog.json.");
 if(operations.length>MAX_OPERATIONS)throw Error("More than 1,000 documented operations.");
 const keys=new Set();
 operations=operations.filter(op=>{const key=[op.method,op.path,op.sourceURL].join("|");if(keys.has(key))return false;keys.add(key);return true;});
 operations.sort((a,b)=>a.path.localeCompare(b.path)||METHODS.indexOf(a.method)-METHODS.indexOf(b.method));
 const paths=new Set(operations.map(x=>x.path));
 const methodCounts=Object.create(null);for(const op of operations)methodCounts[op.method]=(methodCounts[op.method]||0)+1;
 return {
  kind,apiName,version,host,server,operations,documents,
  stats:{operations:operations.length,paths:paths.size,documents:documents.length,methods:methodCounts},
  warnings,networkRequests:0,mode:"DOCUMENTATION_ATLAS_ONLY"
 };
}
export function makeTree(model){
 if(!model||!Array.isArray(model.operations))throw Error("Load a catalog first.");
 const root={id:"/",name:"/",path:"/",depth:0,parent:null,children:[],operations:[]};
 const byPath=new Map([["/",root]]);
 let maxDepth=0;
 for(const op of model.operations){
   let node=root,current="";
   const parts=op.path.split("/").filter(Boolean);
   for(const part of parts){
     current+="/"+part;
     let child=byPath.get(current);
     if(!child){
       child={id:current,name:part,path:current,depth:node.depth+1,parent:node.path,children:[],operations:[]};
       node.children.push(child);byPath.set(current,child);
       maxDepth=Math.max(maxDepth,child.depth);
     }
     node=child;
   }
   node.operations.push(op);
 }
 const nodes=[...byPath.values()].sort((a,b)=>a.depth-b.depth||a.path.localeCompare(b.path));
 const edges=nodes.filter(n=>n.parent).map(n=>({from:n.parent,to:n.id}));
 return {root,nodes,edges,maxDepth,operationCount:model.operations.length};
}
const SECRET_KEYS=/(?:password|passwd|secret|api[-_]?key|access[-_]?token|refresh[-_]?token|private[-_]?key)/i;
function safeSample(value,depth=0){
 if(depth>4)return "<nested value>";
 if(Array.isArray(value))return value.slice(0,10).map(v=>safeSample(v,depth+1));
 if(isObj(value)){
  const out=Object.create(null);
  for(const [key,val] of Object.entries(value).slice(0,40)){
   out[key]=SECRET_KEYS.test(key)?"<REDACTED: enter locally>":safeSample(val,depth+1);
  }
  return out;
 }
 if(typeof value==="string")return value.slice(0,500);
 if(value===null||typeof value==="boolean"||typeof value==="number")return value;
 return "<unsupported sample>";
}
function sampleOf(schema,name="",depth=0){
 if(SECRET_KEYS.test(name))return "<REDACTED: enter locally>";
 if(depth>3)return "<nested object>";
 if(!isObj(schema))return "<"+(name||"value")+">";
 if(schema.$ref)return {unresolvedRef:String(schema.$ref).slice(0,300),warning:"Referenced schema not expanded"};
 if(Array.isArray(schema.enum)&&schema.enum.length)return safeSample(schema.enum[0]);
 if(Object.hasOwn(schema,"example"))return safeSample(schema.example);
 if(Object.hasOwn(schema,"default"))return safeSample(schema.default);
 if(schema.type==="boolean")return false;
 if(schema.type==="integer"||schema.type==="number")return 0;
 if(schema.type==="array")return [sampleOf(schema.items||{},name,depth+1)];
 if(schema.type==="object"||isObj(schema.properties)){
   const out=Object.create(null);
   for(const k of Object.keys(schema.properties||{}).slice(0,20))out[k]=sampleOf(schema.properties[k],k,depth+1);
   return out;
 }
 if(schema.format==="date")return "YYYY-MM-DD";
 if(schema.format==="date-time")return "YYYY-MM-DDTHH:mm:ssZ";
 if(schema.format==="uuid")return "00000000-0000-0000-0000-000000000000";
 return "<"+(name||schema.type||"string")+">";
}
function shq(text){return "'"+String(text).replace(/'/g,"'\\''")+"'";}
export function makeExample(model,op){
 if(!model||!op||!METHODS.includes(op.method))throw Error("Select one documented API operation.");
 const parameters=op.parameters.map(p=>({...p,example:sampleOf(p.schema,p.name)}));
 const host=model.server||"{{baseUrl}}";
 const path=op.path.replace(/\{([^}]+)\}/g,"{{$1}}");
 const url=host.replace(/\/$/,"")+path;
 const headers=[{name:"Accept",value:"application/json"}];
 let body=null,contentType="";
 if(op.request&&isObj(op.request)){
   contentType=String(op.request.content_type||"");
   if(contentType)headers.push({name:"Content-Type",value:contentType});
   if(Object.hasOwn(op.request,"example"))body=safeSample(op.request.example);
   if(op.request.unresolved_ref)body={unresolvedRef:op.request.unresolved_ref,warning:"Reference was not resolved in the source catalog"};
 }
 if(op.requestBody){
   const content=op.requestBody.content;
   if(isObj(content)){
     const type=Object.keys(content).find(x=>x==="application/json")||Object.keys(content)[0];
     if(type){
       contentType=type;headers.push({name:"Content-Type",value:type});
       body=sampleOf(content[type]?.schema||{},"body");
     }
   }
   if(op.requestBody.$ref)body={"$ref":op.requestBody.$ref,"warning":"Unresolved. Do not send until manually reviewed."};
 }
 const lines=["curl --request "+op.method+" \\", "  --url "+shq(url)];
 for(const h of headers)lines.push("  --header "+shq(h.name+": "+h.value));
 for(const p of parameters)if(p.in==="header")lines.push("  # header parameter: "+p.name+" = "+String(p.example).slice(0,100));
 const queries=parameters.filter(p=>p.in==="query").map(p=>p.name+" = "+String(p.example));
 if(queries.length)lines.push("  # query parameters: "+queries.join(", "));
 if(body!==null)lines.push("  --data "+shq(JSON.stringify(body)));
 const curl=lines.join(" \\\n").replace(/ \\\n  #/g,"\n  #");
 return {
   method:op.method,documentedPath:op.path,displayURL:url,operationId:op.operationId,
   parameters,headers,body,contentType,curl,
   source:{url:op.sourceURL,sha256:op.sourceSHA256},
   reviewRequired:!["GET","HEAD"].includes(op.method)||parameters.some(x=>x.required)||op.security.length>0||op.unresolvedRefs,
   caveat:"PREVIEW ONLY. Command is not executed. Placeholders must be supplied, auth verified, target explicitly authorized."
 };
}
/**
 * Assemble editable request EXAMPLES. This never performs HTTP requests.
 * Unfilled required values remain clearly marked as {{placeholders}}.
 */
export function buildRequestPreview(model,op,input={}){
 const example=makeExample(model,op);
 const values=isObj(input.values)?input.values:{};
 let missing=[];
 const path=op.path.replace(/\{([^}]+)\}/g,(_,name)=>{
   const v=values["path:"+name];
   if(v===undefined||v===""){missing.push("path:"+name);return "{{"+name+"}}";}
   return encodeURIComponent(String(v));
 });
 const query=new URLSearchParams();
 const headers=example.headers.map(x=>({...x}));
 for(const p of example.parameters){
   const value=values[p.in+":"+p.name];
   if(p.required&&(value===undefined||value==="")&&p.in!=="path")missing.push(p.in+":"+p.name);
   if(value===undefined||value==="")continue;
   if(p.in==="query")query.append(p.name,String(value));
   else if(p.in==="header"){
     if(!/^[A-Za-z0-9!#$%&'*+.^_~|-]+$/.test(p.name)||/[\r\n]/.test(String(value)))
       throw Error("Unsafe or multiline header parameter: "+p.name);
     headers.push({name:p.name,value:String(value)});
   }
 }
 let url=(model.server||"{{baseUrl}}").replace(/\/$/,"")+path;
 if(query.toString())url+=(url.includes("?")?"&":"?")+query.toString();
 let body=example.body;
 if(typeof input.bodyText==="string"&&input.bodyText.trim()){
   try{body=JSON.parse(input.bodyText);}catch(e){throw Error("Input body must contain valid JSON: "+e.message);}
 }
 const lines=["curl --request "+op.method,"--url "+shq(url)];
 for(const h of headers)lines.push("--header "+shq(h.name+": "+h.value));
 if(body!==null)lines.push("--data "+shq(JSON.stringify(body)));
 const curl=lines.join(" \\\n  ");
 const fetchOptions={method:op.method,headers:Object.fromEntries(headers.map(h=>[h.name,h.value])),credentials:"omit",redirect:"error"};
 if(body!==null)fetchOptions.body=JSON.stringify(body);
 const js="const response = await fetch("+JSON.stringify(url)+", "+JSON.stringify(fetchOptions,null,2)+");\n"+
   "console.log(response.status, await response.text());\n"+
   "// Review the documented method, target, authorization and possible side effects before using.";
 const pyHeaders=Object.fromEntries(headers.map(h=>[h.name,h.value]));
 const python="import json\nimport requests\n"+
   "response = requests.request("+JSON.stringify(op.method)+", "+JSON.stringify(url)+
   ", headers="+JSON.stringify(pyHeaders)+
   (body===null?"":", json=json.loads("+JSON.stringify(JSON.stringify(body))+")")+
   ", timeout=10, allow_redirects=False)\nprint(response.status_code, response.text)\n"+
   "# Illustrative code only. Validate permission and authentication first.";
 return {
  method:op.method,url,headers,body,missing,curl,javascript:js,python,
  readyForReview:missing.length===0&&!op.unresolvedRefs,
  disclaimer:"GENERATED TEXT ONLY. RESTLESS did not execute any API method. Not an authorization or availability check."
 };
}

export function toPostman(model){
 if(!model||!Array.isArray(model.operations))throw Error("Load an API library first.");
 const items=model.operations.map(op=>{
   const example=makeExample(model,op);
   const raw="{{baseUrl}}"+op.path;
   return {
    name:(op.summary||op.operationId||op.method+" "+op.path).slice(0,120),
    request:{
     method:op.method,
     header:example.headers.map(h=>({key:h.name,value:h.value,type:"text"})),
     url:{raw,host:["{{baseUrl}}"],path:op.path.split("/").filter(Boolean),
      query:example.parameters.filter(p=>p.in==="query").map(p=>({key:p.name,value:String(p.example),disabled:true,
       description:"Example only, explicit opt-in before sending"}))},
     description:JSON.stringify({summary:op.summary,sourceURL:op.sourceURL,sourceSHA256:op.sourceSHA256,
       review:"Documentation only; validate domain, authentication and permissions before sending."}),
     ...(example.body!==null?{body:{mode:"raw",raw:JSON.stringify(example.body,null,2),options:{raw:{language:"json"}}}}:{})
    },
    response:[]
   };
 });
 return {
  info:{name:"RESTLESS 42 · "+model.apiName,
   description:"Generated from documented API operations. No endpoints guessed or contacted. Review all calls and parameters before execution.",
   schema:"https://schema.getpostman.com/json/collection/v2.1.0/collection.json"},
  item:items,
  variable:[{key:"baseUrl",value:model.server||"https://YOUR-AUTHORIZED-HOST.example",type:"string"}]
 };
}
export function toMermaid(model){
 const tree=makeTree(model);
 const ids=new Map(tree.nodes.map((n,i)=>[n.id,"n"+i]));
 const esc=x=>String(x).replace(/[<>&\[\]{}"\\\n\r`]/g,"_").replace(/\s+/g," ").slice(0,85);
 const lines=["flowchart LR"];
 for(const n of tree.nodes){
   const label=n.path==="/"?model.apiName:n.name;
   lines.push("  "+ids.get(n.id)+'["'+esc(label)+'"]');
 }
 for(const edge of tree.edges)lines.push("  "+ids.get(edge.from)+" --> "+ids.get(edge.to));
 return lines.join("\n");
}
