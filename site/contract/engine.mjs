/**
 * RESTLESS 42 Contract Lab.
 * Analysis of locally pasted/imported OpenAPI 3 JSON. No network access, no code execution.
 * Heuristic compatibility signals are NOT evidence of deployed behavior or semantic compatibility.
 */
import {parseOpenAPIOperations,searchOpenAPIOperations} from "../toolbox/transforms.mjs";

const MAX_CHARS=2_000_000, MAX_OPS=1000;
const READ_METHODS=new Set(["GET","HEAD","OPTIONS"]);
const SORT_KEY=(a,b)=>a.key.localeCompare(b.key);
function checkSource(source){
  if(typeof source!=="string"||source.length>MAX_CHARS)throw Error("Expected an OpenAPI JSON document of at most 2 million characters.");
}
function parseSpec(source){
  checkSource(source);
  let spec;
  try{spec=JSON.parse(source);}catch(e){throw Error("Invalid OpenAPI JSON: "+e.message);}
  if(!spec||typeof spec!=="object"||Array.isArray(spec)||typeof spec.openapi!=="string"||!spec.openapi.startsWith("3."))throw Error("Expected an OpenAPI 3.x JSON document.");
  return spec;
}
function normalizeSecurity(security){
  if(!Array.isArray(security))return [];
  return security.map(item=>{
    if(!item||typeof item!=="object"||Array.isArray(item))return {};
    const out={};
    for(const key of Object.keys(item).sort()){
      out[key]=Array.isArray(item[key])?[...item[key]].sort():item[key];
    }
    return out;
  }).sort((a,b)=>JSON.stringify(a).localeCompare(JSON.stringify(b)));
}
function normalizeOperation(op){
  return {
    operationId:op.operationId,
    summary:op.summary,
    description:op.description,
    tags:[...op.tags].sort(),
    parameters:op.parameters.map(x=>({
      name:x.name,in:x.in,required:!!x.required,description:x.description||"",
      ref:x.ref||"",unresolvedRef:!!x.unresolvedRef
    })).sort((a,b)=>(a.in+":"+a.name).localeCompare(b.in+":"+b.name)),
    security:normalizeSecurity(op.security),
    unresolvedRefs:!!op.unresolvedRefs
  };
}
function opKey(op){return op.method.toUpperCase()+" "+op.path;}
function indexOperations(operations){
  const out=new Map();
  for(const op of operations){
    const key=opKey(op);
    if(out.has(key))throw Error("Duplicate documented operation: "+key);
    out.set(key,op);
  }
  return out;
}
function title(spec){return typeof spec.info?.title==="string"?spec.info.title:"Untitled API";}
function specVersion(spec){return typeof spec.info?.version==="string"?spec.info.version:"unversioned";}
function methodCounts(ops){
  const out=Object.create(null);
  for(const op of ops)out[op.method]=(out[op.method]||0)+1;
  return out;
}
function lintSpec(spec,ops){
  const findings=[];
  if(!spec.info?.title)findings.push({severity:"info",code:"MISSING_TITLE",message:"No info.title is declared."});
  if(!spec.info?.version)findings.push({severity:"info",code:"MISSING_VERSION",message:"No info.version is declared."});
  if(!spec.servers||!spec.servers.length)findings.push({severity:"info",code:"NO_SERVERS",message:"No servers list is declared; RESTLESS does not infer or probe a host."});
  for(const op of ops){
    const key=opKey(op);
    const inPath=[...op.path.matchAll(/\{([^{}]+)\}/g)].map(x=>x[1]);
    for(const name of inPath){
      if(!op.parameters.some(p=>p.in==="path"&&p.name===name&&p.required))
        findings.push({severity:"review",code:"UNDECLARED_PATH_PARAMETER",operation:key,message:"Path placeholder {"+name+"} lacks an explicit required path parameter."});
    }
    if(op.unresolvedRefs)findings.push({severity:"review",code:"UNRESOLVED_REF",operation:key,message:"Reference resolution is incomplete; review the source specification."});
    if(!op.operationId)findings.push({severity:"info",code:"NO_OPERATION_ID",operation:key,message:"No operationId; this is allowed by OpenAPI."});
    if(!READ_METHODS.has(op.method)&&op.security.length===0)
      findings.push({severity:"info",code:"NO_DECLARED_SECURITY",operation:key,message:"This mutation operation does not declare security requirements. Absence is not proof of public access."});
    if(findings.length>=250){findings.push({severity:"info",code:"FINDINGS_LIMIT",message:"The report shows at most 250 findings."});break;}
  }
  return findings;
}
export function inspectContract(source){
  const spec=parseSpec(source),operations=parseOpenAPIOperations(spec);
  if(operations.length>MAX_OPS)throw Error("Operation limit exceeded.");
  const counts=methodCounts(operations);
  const securityDeclared=operations.filter(x=>x.security.length>0).length;
  const needsReview=operations.filter(x=>x.reviewRequired||x.unresolvedRefs).length;
  const paths=new Set(operations.map(op=>op.path));
  const tags=new Set(operations.flatMap(op=>op.tags));
  return {
    mode:"DOCUMENTATION_ONLY_NO_EXECUTION",
    networkRequests:0,format:"OpenAPI 3 JSON",
    api:{title:title(spec),version:specVersion(spec),openapi:spec.openapi},
    stats:{operations:operations.length,uniquePaths:paths.size,tagCount:tags.size,securityDeclared,needsReview,methods:counts},
    findings:lintSpec(spec,operations),
    operations:operations.map(op=>({...op,key:opKey(op),reviewRequired:op.reviewRequired||op.unresolvedRefs}))
  };
}
function changeReasons(before,after){
  const reasons=[];
  const b=normalizeOperation(before),a=normalizeOperation(after);
  if(b.operationId!==a.operationId)reasons.push("operationId");
  if(b.summary!==a.summary)reasons.push("summary");
  if(b.description!==a.description)reasons.push("description");
  if(JSON.stringify(b.tags)!==JSON.stringify(a.tags))reasons.push("tags");
  if(JSON.stringify(b.parameters)!==JSON.stringify(a.parameters))reasons.push("parameters");
  if(JSON.stringify(b.security)!==JSON.stringify(a.security))reasons.push("declared security");
  if(b.unresolvedRefs!==a.unresolvedRefs)reasons.push("reference resolution");
  return reasons;
}
function changeRisk(before,after,reasons){
  const oldParams=new Map(before.parameters.map(p=>[p.in+":"+p.name,p]));
  const changedRequired=after.parameters.some(p=>p.required&&(!oldParams.has(p.in+":"+p.name)||!oldParams.get(p.in+":"+p.name).required));
  const removedParam=before.parameters.some(p=>!after.parameters.some(q=>q.name===p.name&&q.in===p.in));
  const authIncrease=before.security.length===0&&after.security.length>0;
  if(changedRequired||removedParam||authIncrease)return {level:"potentially-breaking",notes:[
    ...(changedRequired?["Newly required parameter detected"]:[]),
    ...(removedParam?["Previously documented parameter removed"]:[]),
    ...(authIncrease?["Security requirements newly declared"]:[])
  ]};
  if(reasons.length)return {level:"review",notes:["Documentation contract changed; compatibility requires manual analysis"]};
  return {level:"none",notes:[]};
}
export function compareContracts(beforeText,afterText){
  const before=inspectContract(beforeText),after=inspectContract(afterText);
  const older=indexOperations(before.operations),newer=indexOperations(after.operations);
  const keys=[...new Set([...older.keys(),...newer.keys()])].sort();
  const changes=[];
  const counts={added:0,removed:0,changed:0,unchanged:0,potentiallyBreaking:0};
  for(const key of keys){
    const old=older.get(key),current=newer.get(key);
    if(!old){counts.added++;changes.push({key,type:"added",risk:"review",before:null,after:current,reasons:["New documented operation"]});continue;}
    if(!current){counts.removed++;counts.potentiallyBreaking++;changes.push({key,type:"removed",risk:"potentially-breaking",before:old,after:null,reasons:["Previously documented operation was removed"]});continue;}
    const reasons=changeReasons(old,current);
    if(reasons.length){
      const risk=changeRisk(old,current,reasons);
      counts.changed++;if(risk.level==="potentially-breaking")counts.potentiallyBreaking++;
      changes.push({key,type:"changed",risk:risk.level,before:old,after:current,reasons:[...reasons,...risk.notes]});
    }else counts.unchanged++;
  }
  changes.sort(SORT_KEY);
  return {
    mode:"OFFLINE_CONTRACT_DIFF",networkRequests:0,
    baseline:before.api,candidate:after.api,counts,changes,
    notes:["Only documented method and path identify an operation.",
      "These are heuristic contract-level signals, not verified breaking changes or live server checks.",
      "Changed source URLs and server deployment status are not compared.",
      "No HTTP request, authentication attempt, or execution occurs."]
  };
}
export function findOperations(report,search="",method="ALL",onlyReview=false){
  const query=search.trim();
  let items=report.operations;
  if(query)items=searchOpenAPIOperations(items,query,1000);
  return items.filter(op=>(method==="ALL"||op.method===method)&&(!onlyReview||op.reviewRequired));
}
