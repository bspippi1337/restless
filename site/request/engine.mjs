/**
 * RESTLESS 42 Request Lab.
 * Opt-in, one-request browser HTTP inspection. NOT the offline Toolbox.
 * No scanning, mutation, retries or redirect following.
 */
export const METHODS=Object.freeze(["GET","HEAD","OPTIONS"]);
const MAX_RESPONSE_BYTES=2_000_000;
const MAX_HEADER_BYTES=8192;
const MAX_HEADERS=20;
const BLOCKED_HEADERS=new Set(["host","cookie","set-cookie","origin","referer","content-length",
 "connection","accept-encoding","transfer-encoding","upgrade","user-agent",
 "proxy-authorization","proxy-authenticate"]);
function forbiddenHost(host){
 const lower=host.toLowerCase().replace(/\.$/,"");
 if(!lower.includes(".")||lower.includes(":"))return true;
 if(["localhost","localhost.localdomain"].includes(lower))return true;
 for(const suffix of [".localhost",".local",".internal",".home",".lan",".onion",".invalid",".test",".example"])
   if(lower.endsWith(suffix))return true;
 if(/^\d{1,3}(?:\.\d{1,3}){3}$/.test(lower))return true;
 if(!/^[a-z0-9.-]+$/.test(lower)||lower.startsWith(".")||lower.includes(".."))return true;
 return false;
}
function parseHeaders(input){
 if(typeof input!=="string"||input.length>MAX_HEADER_BYTES)throw Error("Custom headers must be a JSON object under 8 KB.");
 let object;
 try{object=input.trim()?JSON.parse(input):{};}catch(e){throw Error("Invalid headers JSON: "+e.message);}
 if(!object||typeof object!=="object"||Array.isArray(object))throw Error("Custom headers must be a JSON object.");
 const entries=Object.entries(object);
 if(entries.length>MAX_HEADERS)throw Error("Maximum 20 custom headers.");
 const headers=new Headers();
 for(const [name,value]of entries){
   if(typeof value!=="string")throw Error("Header values must be strings: "+name);
   if(value.length>2000)throw Error("Header value is too long: "+name);
   if(!/^[!#$%&'*+.^_~0-9A-Za-z-]+$/.test(name))throw Error("Invalid header name: "+name);
   const lower=name.toLowerCase();
   if(BLOCKED_HEADERS.has(lower)||lower.startsWith("sec-")||lower.startsWith("proxy-"))throw Error("Browser-controlled header is not permitted: "+name);
   if(/[\r\n]/.test(value))throw Error("Header value cannot contain line breaks: "+name);
   headers.set(name,value);
 }
 return headers;
}
export function validateRequest(config){
 const {url,allowedHost,authorized,method="GET",headersJson="{}",allowAuthorization=false,timeoutMs=10000}=config||{};
 if(authorized!==true)throw Error("Explicit authorization confirmation is required for every request.");
 if(typeof url!=="string"||url.length>2048)throw Error("Provide a URL no longer than 2,048 characters.");
 let target;
 try{target=new URL(url.trim());}catch{throw Error("Enter a valid absolute HTTPS URL.");}
 if(target.protocol!=="https:")throw Error("Request Lab only supports HTTPS URLs.");
 if(target.username||target.password||target.hash)throw Error("URL credentials and fragments are not allowed.");
 if(forbiddenHost(target.hostname))throw Error("Request Lab requires a public DNS hostname (not an IP literal, localhost or reserved name).");
 if(typeof allowedHost!=="string"||!allowedHost.trim())throw Error("Enter the target hostname separately in the allowlist field.");
 const expected=allowedHost.trim().toLowerCase().replace(/\.$/,"");
 if(target.hostname.toLowerCase().replace(/\.$/,"")!==expected)throw Error("Target hostname does not match the approved host.");
 if(!METHODS.includes(method))throw Error("Only GET, HEAD and OPTIONS are supported.");
 if(!Number.isInteger(timeoutMs)||timeoutMs<1000||timeoutMs>15000)throw Error("Timeout must be 1–15 seconds.");
 const headers=parseHeaders(headersJson);
 if(headers.has("Authorization")&&!allowAuthorization)throw Error("Authorization header requires separate credential confirmation.");
 return {url:target.href,host:target.hostname,method,headers,timeoutMs,
   consent:authorized,allowAuthorization:!!allowAuthorization};
}
function hex(bytes){return Array.from(bytes,b=>b.toString(16).padStart(2,"0")).join("");}
async function readBounded(response){
 if(!response.body)return new Uint8Array();
 const reader=response.body.getReader();let size=0,chunks=[];
 try{
   while(true){
     const {value,done}=await reader.read();
     if(done)break;
     if(!(value instanceof Uint8Array))throw Error("Unsupported response stream type.");
     size+=value.byteLength;
     if(size>MAX_RESPONSE_BYTES)throw Error("Response exceeded the 2 MB safety limit; read stopped.");
     chunks.push(value);
   }
 }catch(error){try{await reader.cancel();}catch{}throw error;}
 finally{try{reader.releaseLock();}catch{}}
 const bytes=new Uint8Array(size);let offset=0;
 for(const chunk of chunks){bytes.set(chunk,offset);offset+=chunk.byteLength;}
 return bytes;
}
export async function executeRequest(config,transport=globalThis.fetch){
 const verified=validateRequest(config);
 if(typeof transport!=="function")throw Error("No fetch implementation available.");
 const controller=new AbortController();const timer=setTimeout(()=>controller.abort(),verified.timeoutMs);
 const start=performance.now();
 try{
   const response=await transport(verified.url,{
     method:verified.method,headers:verified.headers,
     signal:controller.signal,cache:"no-store",mode:"cors",
     credentials:"omit",redirect:"error",referrerPolicy:"no-referrer"
   });
   const declared=Number(response.headers.get("content-length"));
   if(Number.isFinite(declared)&&declared>MAX_RESPONSE_BYTES)throw Error("Response Content-Length exceeds the 2 MB safety limit.");
   const bytes=await readBounded(response);
   const durationMs=Math.round(performance.now()-start);
   const observedHeaders={};
   for(const [name,value]of response.headers.entries())observedHeaders[name]=value;
   let body,bodyFormat="text";
   try{
     body=new TextDecoder("utf-8",{fatal:true}).decode(bytes);
     if(response.headers.get("content-type")?.toLowerCase().includes("json")){
       try{body=JSON.stringify(JSON.parse(body),null,2);bodyFormat="formatted-json";}
       catch{bodyFormat="text-invalid-json";}
     }
   }catch{body="Binary response, hex prefix:\n"+hex(bytes.slice(0,512));bodyFormat="binary-preview";}
   const digest=await crypto.subtle.digest("SHA-256",bytes);
   return{
     mode:"ONE_EXPLICIT_AUTHORIZED_BROWSER_REQUEST",
     networkRequests:1,method:verified.method,url:verified.url,
     responseStatus:response.status,responseStatusText:response.statusText,
     durationMs,bytesRead:bytes.length,sha256:hex(new Uint8Array(digest)),
     responseHeaders:observedHeaders,responseBody:body,bodyFormat,
     caveats:["No requests are made until the operator approves a target and presses Send.",
       "Browser CORS and Private Network Access policy may prevent cross-origin calls.",
       "HTTP GET may still have effects; access rights and scope remain the user's responsibility.",
       "Response headers reflect what the browser exposes to JavaScript.",
       "Credentials are omitted unless Authorization is explicitly enabled."]
   };
 }catch(error){
   if(controller.signal.aborted)throw Error("Request timed out or was aborted after "+verified.timeoutMs+" ms.");
   if(error instanceof TypeError)throw Error("Browser request failed. The host may be unavailable, TLS invalid, or CORS may forbid this origin. No bypass is attempted.");
   throw error;
 }finally{clearTimeout(timer);}
}
