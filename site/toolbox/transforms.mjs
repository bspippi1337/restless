// RESTLESS 42 toolbox. Pure local data transforms; no fetch, eval, storage or remote scripts.
export const OPERATIONS = Object.freeze([
{id:"json-pretty",label:"JSON: pretty print",group:"JSON",description:"Format validated JSON with two-space indentation."},
{id:"json-minify",label:"JSON: minify",group:"JSON",description:"Validate JSON and remove whitespace."},
{id:"json-sort",label:"JSON: sort keys",group:"JSON",description:"Recursively sort object keys."},
{id:"json-path",label:"JSON: extract path",group:"JSON",option:"Path",defaultOption:"items[0].name",description:"Extract a dotted path with [0] array indexes."},
{id:"json-to-csv",label:"JSON array to CSV",group:"JSON",description:"Convert an array of objects to properly escaped CSV."},
{id:"csv-to-json",label:"CSV to JSON",group:"JSON",description:"Parse CSV with multiline quoted values."},
{id:"base64-encode",label:"To Base64",group:"Encoding",description:"Encode UTF-8 text as Base64."},
{id:"base64-decode",label:"From Base64",group:"Encoding",description:"Decode Base64 into UTF-8 text."},
{id:"base64url-encode",label:"To Base64URL",group:"Encoding",description:"Encode UTF-8 using URL-safe Base64 without padding."},
{id:"base64url-decode",label:"From Base64URL",group:"Encoding",description:"Decode URL-safe Base64 as UTF-8."},
{id:"url-encode",label:"URL component encode",group:"Encoding",description:"Percent-encode one URL component."},
{id:"url-decode",label:"URL component decode",group:"Encoding",description:"Decode URL component percent escapes."},
{id:"hex-encode",label:"UTF-8 to Hex",group:"Encoding",description:"Get hexadecimal UTF-8 bytes."},
{id:"hex-decode",label:"Hex to UTF-8",group:"Encoding",description:"Decode hexadecimal UTF-8 bytes."},
{id:"html-escape",label:"Escape HTML",group:"Encoding",description:"Escape &, <, >, double quotes and apostrophes."},
{id:"html-unescape",label:"Unescape HTML",group:"Encoding",description:"Decode the five common HTML entities."},
{id:"jwt-inspect",label:"JWT: inspect unverified",group:"Security",description:"Read JWT header/payload. Does not verify signatures or claims."},
{id:"sha-256",label:"SHA-256 digest",group:"Security",description:"Hash UTF-8 text using Web Crypto."},
{id:"sha-512",label:"SHA-512 digest",group:"Security",description:"Hash UTF-8 text using Web Crypto."},
{id:"rot13",label:"ROT13",group:"Text",description:"Reversible alphabet substitution."},
{id:"uppercase",label:"Uppercase",group:"Text",description:"Make text uppercase."},
{id:"lowercase",label:"Lowercase",group:"Text",description:"Make text lowercase."},
{id:"trim",label:"Trim whitespace",group:"Text",description:"Remove leading/trailing whitespace."},
{id:"sort-lines",label:"Sort lines",group:"Text",description:"Sort lines lexically."},
{id:"unique-lines",label:"Unique lines",group:"Text",description:"Remove repeated lines, preserving order."},
{id:"text-stats",label:"Text statistics",group:"Text",description:"Count Unicode characters, bytes, words and lines."},
{id:"url-inspect",label:"Inspect URL",group:"Web",description:"Parse URL components locally."},
{id:"query-parse",label:"Query string to JSON",group:"Web",description:"Parse all query keys, preserving duplicates."},
{id:"query-stringify",label:"JSON to query string",group:"Web",description:"Serialize an object using percent encoding."},
{id:"openapi-operations",label:"OpenAPI: list operations",group:"OpenAPI",description:"Extract documented HTTP operations from OpenAPI 3 JSON."},
{id:"openapi-search",label:"OpenAPI: find operation",group:"OpenAPI",option:"Keywords",defaultOption:"list pets",description:"Rank documented operations via explainable lexical scoring."},
{id:"openapi-plan",label:"OpenAPI: offline probe plan",group:"OpenAPI",option:"Task",defaultOption:"find pet",description:"Create a review-only plan without making HTTP requests."}
]);
export const OPERATION_BY_ID=new Map(OPERATIONS.map(op=>[op.id,op]));
const encoder=new TextEncoder(),decoder=new TextDecoder("utf-8",{fatal:true});
const MAX_CHARS=2_000_000,MAX_STEPS=24;
const METHODS=new Set(["get","head","options","post","put","patch","delete","trace"]);
function checkText(text){if(typeof text!=="string")throw Error("Input must be text.");if(text.length>MAX_CHARS)throw Error("Input exceeds 2 million characters.");return text;}
function pretty(x){return JSON.stringify(x,null,2);}
function parseJSON(s){try{return JSON.parse(s);}catch(e){throw Error("Invalid JSON: "+e.message);}}
function hex(bytes){return [...bytes].map(x=>x.toString(16).padStart(2,"0")).join("");}
function fromHex(text){const s=text.replace(/\s/g,"");if(!/^(?:[0-9a-fA-F]{2})*$/.test(s))throw Error("Hex data must consist of pairs of digits.");return Uint8Array.from(s.match(/.{2}/g)||[],x=>parseInt(x,16));}
function base64(bytes){let binary="";for(let i=0;i<bytes.length;i+=8192)binary+=String.fromCharCode(...bytes.slice(i,i+8192));return btoa(binary);}
function fromBase64(text,url=false){let s=text.replace(/\s/g,"");if(url)s=s.replace(/-/g,"+").replace(/_/g,"/");if(!/^[A-Za-z0-9+/]*={0,2}$/.test(s)||s.length%4===1)throw Error("Invalid Base64 characters or padding.");s=s.padEnd(Math.ceil(s.length/4)*4,"=");return Uint8Array.from(atob(s),c=>c.charCodeAt(0));}
function sortJSON(value){if(Array.isArray(value))return value.map(sortJSON);if(value&&typeof value==="object"){const out={};for(const k of Object.keys(value).sort())out[k]=sortJSON(value[k]);return out;}return value;}
function extractPath(text,path){
  if(!path.trim())throw Error("Enter a property path, e.g. items[0].name.");
  const normalized=path.replace(/^\$\.?/,""),parts=normalized.match(/[^.[\]]+|\[\d+\]/g);
  if(!parts||parts.join("").replace(/[\[\]]/g,"")!==normalized.replace(/[.\[\]]/g,""))throw Error("Use simple dotted properties and [0] numeric indexes.");
  let v=parseJSON(text);
  for(const raw of parts){const key=raw.startsWith("[")?Number(raw.slice(1,-1)):raw;if(v==null||!Object.prototype.hasOwnProperty.call(v,key))throw Error("Path not found: "+path);v=v[key];}
  return typeof v==="string"?v:pretty(v);
}
function csvCell(value){const s=value==null?"":typeof value==="object"?JSON.stringify(value):String(value);return /[",\r\n]/.test(s)?'"'+s.replace(/"/g,'""')+'"':s;}
function toCSV(input){const rows=parseJSON(input);if(!Array.isArray(rows)||!rows.every(x=>x&&typeof x==="object"&&!Array.isArray(x)))throw Error("Expected a JSON array of objects.");const keys=[...new Set(rows.flatMap(Object.keys))];if(!keys.length)return "";return[keys.map(csvCell).join(","),...rows.map(r=>keys.map(k=>csvCell(r[k])).join(","))].join("\r\n");}
function parseCSV(text){let rows=[],row=[],cell="",quoted=false,justClosed=false;
  for(let i=0;i<text.length;i++){const ch=text[i];
    if(quoted){if(ch==='"'&&text[i+1]==='"'){cell+='"';i++;}else if(ch==='"'){quoted=false;justClosed=true;}else cell+=ch;continue;}
    if(justClosed){if(ch!==","&&ch!=="\r"&&ch!=="\n")throw Error("Invalid content after closing CSV quote.");justClosed=false;}
    if(ch==='"'&&cell===""){quoted=true;continue;}
    if(ch==='"')throw Error("Unexpected quote inside CSV field.");
    if(ch===","){row.push(cell);cell="";continue;}
    if(ch==="\r"||ch==="\n"){row.push(cell);rows.push(row);row=[];cell="";if(ch==="\r"&&text[i+1]==="\n")i++;if(rows.length>5001)throw Error("CSV exceeds 5,000 rows.");continue;}
    cell+=ch;
  }
  if(quoted)throw Error("Unterminated CSV quoted field.");
  if(cell!==""||row.length){row.push(cell);rows.push(row);}
  while(rows.length&&rows[rows.length-1].length===1&&rows[rows.length-1][0]==="")rows.pop();
  return rows;
}
function fromCSV(text){const rows=parseCSV(text);if(!rows.length)return"[]";const keys=rows.shift();if(new Set(keys).size!==keys.length)throw Error("Duplicate CSV header.");return pretty(rows.map((r,i)=>{if(r.length!==keys.length)throw Error("CSV row "+(i+2)+" has "+r.length+" cells, expected "+keys.length);return Object.fromEntries(keys.map((key,j)=>[key,r[j]]));}));}
const escapeMap={"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"};
const decodeMap={"&amp;":"&","&lt;":"<","&gt;":">","&quot;":'"',"&#39;":"'","&apos;":"'"};
function jwt(text){const parts=text.trim().split(".");if(parts.length!==3)throw Error("JWT must contain three sections.");const head=parseJSON(decoder.decode(fromBase64(parts[0],true))),payload=parseJSON(decoder.decode(fromBase64(parts[1],true)));return pretty({header:head,payload,signaturePresent:!!parts[2],signatureVerified:false,warning:"Unverified inspection only. Do not trust the decoded claims."});}
function urlInspect(text){let u;try{u=new URL(text.trim());}catch{throw Error("Expected an absolute URL including scheme.");}const q={};for(const [k,v]of u.searchParams){if(!q[k])q[k]=[];q[k].push(v);}return pretty({protocol:u.protocol,credentialsPresent:!!(u.username||u.password),host:u.host,hostname:u.hostname,port:u.port,pathname:u.pathname,query:q,fragment:u.hash,note:"Parsed locally. No HTTP request."});}
function queryParse(text){const o={};for(const[k,v]of new URLSearchParams(text.replace(/^\?/,""))){if(!o[k])o[k]=[];o[k].push(v);}return pretty(o);}
function queryStringify(text){const obj=parseJSON(text);if(!obj||typeof obj!=="object"||Array.isArray(obj))throw Error("Expected a JSON object.");const p=new URLSearchParams();for(const[k,v]of Object.entries(obj)){for(const item of Array.isArray(v)?v:[v])p.append(k,item==null?"":typeof item==="object"?JSON.stringify(item):String(item));}return p.toString();}
function words(text){return [...new Set(String(text).toLowerCase().match(/[a-z0-9]+/g)||[])].filter(x=>x.length>1);}
function resolveLocalParameter(param,spec){if(!param?.$ref)return param;const seen=new Set();let ref=param.$ref;for(let i=0;i<8;i++){const prefix="#/components/parameters/";if(!ref.startsWith(prefix)||seen.has(ref))return {...param,unresolvedRef:true};seen.add(ref);const name=ref.slice(prefix.length).replace(/~1/g,"/").replace(/~0/g,"~");const target=spec.components?.parameters?.[name];if(!target)return {...param,unresolvedRef:true};if(!target.$ref)return target;ref=target.$ref;}return {...param,unresolvedRef:true};}
export function parseOpenAPIOperations(input){
  const spec=typeof input==="string"?parseJSON(input):input;
  if(!spec||typeof spec!=="object"||Array.isArray(spec)||typeof spec.openapi!=="string"||!spec.openapi.startsWith("3."))throw Error("Expected OpenAPI 3.x JSON. YAML and Swagger 2 are not supported.");
  if(!spec.paths||typeof spec.paths!=="object"||Array.isArray(spec.paths))throw Error("OpenAPI paths object is required.");
  const ops=[];
  for(const[path,item]of Object.entries(spec.paths)){if(!path.startsWith("/")||!item||typeof item!=="object"||item.$ref)continue;
    for(const[method,raw]of Object.entries(item)){if(!METHODS.has(method)||!raw||typeof raw!=="object")continue;
      const inherited=Array.isArray(item.parameters)?item.parameters:[],defined=Array.isArray(raw.parameters)?raw.parameters:[],parameters=new Map();
      for(const par of [...inherited,...defined]){const p=resolveLocalParameter(par,spec);parameters.set(p?.in&&p?.name?p.in+":"+p.name:JSON.stringify(p),p);}
      const security=Array.isArray(raw.security)?raw.security:Array.isArray(spec.security)?spec.security:[];
      const entries=[...parameters.values()];
      ops.push({method:method.toUpperCase(),path,operationId:String(raw.operationId||""),summary:String(raw.summary||""),description:String(raw.description||"").slice(0,800),tags:Array.isArray(raw.tags)?raw.tags.map(String):[],parameters:entries.map(p=>({name:p?.name||"",in:p?.in||"",required:!!p?.required,description:p?.description||"",ref:p?.$ref||"",unresolvedRef:!!p?.unresolvedRef})),security,reviewRequired:!(["get","head"].includes(method))||security.length>0||path.includes("{")||entries.some(p=>p?.required||p?.unresolvedRef)||!!raw.requestBody?.$ref,unresolvedRefs:entries.some(p=>p?.unresolvedRef)||!!raw.requestBody?.$ref});
      if(ops.length>1000)throw Error("Operation limit exceeded (1,000).");
    }
  }
  return ops.sort((a,b)=>a.path.localeCompare(b.path)||a.method.localeCompare(b.method));
}
export function searchOpenAPIOperations(ops,query,max=20){
  const wanted=words(query);if(!wanted.length)throw Error("Enter operation search keywords.");
  return ops.map(op=>{let score=0;const matched=[];
    for(const[label,content,weight]of [["operation_id",op.operationId,8],["path",op.path,6],["summary",op.summary,6],["description",op.description,2],["tags",op.tags.join(" "),4],["method",op.method,3],["parameters",op.parameters.map(p=>p.name).join(" "),2]]){
      const terms=words(content),matches=wanted.filter(w=>terms.some(t=>t===w||(w.length>3&&t.startsWith(w)))).length;
      if(matches){score+=weight*matches;matched.push(label);}
    }
    if(["GET","HEAD"].includes(op.method))score+=4;
    return {score,matchedFields:matched,...op};
  }).filter(x=>x.matchedFields.length).sort((a,b)=>b.score-a.score||a.path.localeCompare(b.path)||a.method.localeCompare(b.method)).slice(0,max);
}
function planOpenAPI(text,query){const matches=searchOpenAPIOperations(parseOpenAPIOperations(text),query,10);return pretty({mode:"DOCUMENTATION_ONLY_NO_EXECUTION",networkRequests:0,query,source:"Pasted OpenAPI 3 JSON; live endpoint status unverified",candidates:matches.map((m,i)=>({rank:i+1,method:m.method,documentedPath:m.path,operationId:m.operationId,summary:m.summary,relevanceScore:m.score,parameters:m.parameters,declaredSecurity:m.security,requiresManualReview:m.reviewRequired||m.unresolvedRefs,reviewReasons:[...(!["GET","HEAD"].includes(m.method)?["Non-read method"]:[]),...(m.security.length?["Declared authentication"]:[]),...(m.path.includes("{")?["Path parameters"]:[]),...(m.parameters.some(p=>p.required)?["Required parameters"]:[]),...(m.unresolvedRefs?["Unresolved references"]:[])]})),notes:["Offline plan: no requests were made.","A documented operation is not proof of live availability.","Explicit target authorization is required before any live call."]});}
export async function applyOperation(id,input,option=""){
  checkText(input);
  switch(id){
    case"json-pretty":return pretty(parseJSON(input));
    case"json-minify":return JSON.stringify(parseJSON(input));
    case"json-sort":return pretty(sortJSON(parseJSON(input)));
    case"json-path":return extractPath(input,option);
    case"json-to-csv":return toCSV(input);
    case"csv-to-json":return fromCSV(input);
    case"base64-encode":return base64(encoder.encode(input));
    case"base64-decode":return decoder.decode(fromBase64(input));
    case"base64url-encode":return base64(encoder.encode(input)).replace(/\+/g,"-").replace(/\//g,"_").replace(/=+$/,"");
    case"base64url-decode":return decoder.decode(fromBase64(input,true));
    case"url-encode":return encodeURIComponent(input);
    case"url-decode":return decodeURIComponent(input);
    case"hex-encode":return hex(encoder.encode(input));
    case"hex-decode":return decoder.decode(fromHex(input));
    case"html-escape":return input.replace(/[&<>"']/g,x=>escapeMap[x]);
    case"html-unescape":return input.replace(/&(?:amp|lt|gt|quot|apos|#39);/g,x=>decodeMap[x]);
    case"jwt-inspect":return jwt(input);
    case"sha-256":case"sha-512":{if(!globalThis.crypto?.subtle)throw Error("Web Crypto requires HTTPS or localhost.");const digest=await globalThis.crypto.subtle.digest(id==="sha-256"?"SHA-256":"SHA-512",encoder.encode(input));return hex(new Uint8Array(digest));}
    case"rot13":return input.replace(/[A-Za-z]/g,c=>String.fromCharCode(c.charCodeAt(0)+(c.toLowerCase()<="m"?13:-13)));
    case"uppercase":return input.toUpperCase();
    case"lowercase":return input.toLowerCase();
    case"trim":return input.trim();
    case"sort-lines":return input.split(/\r?\n/).sort().join("\n");
    case"unique-lines":return [...new Set(input.split(/\r?\n/))].join("\n");
    case"text-stats":return pretty({characters:[...input].length,utf8Bytes:encoder.encode(input).length,words:input.trim()?input.trim().split(/\s+/).length:0,lines:input===""?0:input.split(/\r\n|\n|\r/).length});
    case"url-inspect":return urlInspect(input);
    case"query-parse":return queryParse(input);
    case"query-stringify":return queryStringify(input);
    case"openapi-operations":return pretty(parseOpenAPIOperations(input));
    case"openapi-search":return pretty(searchOpenAPIOperations(parseOpenAPIOperations(input),option));
    case"openapi-plan":return planOpenAPI(input,option);
    default:throw Error("Unknown operation: "+id);
  }
}
export function validateRecipe(recipe){
 if(!Array.isArray(recipe)||recipe.length>24)throw Error("Expected a recipe with at most 24 operations.");
 return recipe.map((op,i)=>{if(!op||!OPERATION_BY_ID.has(op.id))throw Error("Unknown operation at step "+(i+1));if(op.option!=null&&typeof op.option!=="string")throw Error("Invalid option at step "+(i+1));if(op.option?.length>500)throw Error("Operation settings exceed 500 characters.");return {id:op.id,enabled:op.enabled!==false,option:op.option??OPERATION_BY_ID.get(op.id).defaultOption??""};});
}
export async function runRecipe(input,recipe){
 let output=checkText(input),history=[];
 for(const[i,step]of validateRecipe(recipe).entries()){if(!step.enabled)continue;try{output=await applyOperation(step.id,output,step.option);if(typeof output!=="string")throw Error("Invalid output type.");checkText(output);history.push({step:i+1,id:step.id,characters:output.length});}catch(e){const err=new Error("Step "+(i+1)+" ("+OPERATION_BY_ID.get(step.id).label+"): "+e.message);err.step=i;throw err;}}
 return {output,history};
}
