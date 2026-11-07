import {executeRequest,validateRequest} from "./engine.mjs";
const $=id=>document.getElementById(id);
let busy=false,last=null;
const status=$("status"),response=$("response");
function setStatus(message,kind="info"){
 status.textContent=message;status.dataset.kind=kind;
}
function config(){
 return{
   url:$("url").value,
   allowedHost:$("approved-host").value,
   authorized:$("permission").checked,
   allowAuthorization:$("credentials").checked,
   method:$("method").value,
   timeoutMs:Number($("timeout").value),
   headersJson:$("headers").value
 };
}
function updateSendState(){
 $("send").disabled=busy||!$("permission").checked||!$("url").value.trim()||!$("approved-host").value.trim();
}
function renderResult(result){
 last=result;response.hidden=false;
 $("http-status").textContent=result.responseStatus+" "+result.responseStatusText;
 $("duration").textContent=result.durationMs+" ms";
 $("body-size").textContent=result.bytesRead.toLocaleString()+" bytes";
 $("sha").textContent=result.sha256;
 $("body").value=result.responseBody;
 $("response-headers").value=JSON.stringify(result.responseHeaders,null,2);
}
async function send(){
 if(busy)return;
 let cfg;try{cfg=validateRequest(config());}catch(e){setStatus(e.message,"error");return;}
 busy=true;updateSendState();last=null;response.hidden=true;
 setStatus("Sending ONE explicitly approved "+cfg.method+" request to "+cfg.host+". No retries or redirects.");
 try{
   const start=performance.now();
   const result=await executeRequest(config());
   renderResult(result);
   setStatus("Received HTTP "+result.responseStatus+" in "+result.durationMs+" ms; "+result.bytesRead+" bytes verified with SHA-256. Network requests: 1.","success");
   response.scrollIntoView({behavior:"smooth",block:"start"});
 }catch(e){
   setStatus(e.message||String(e),"error");
 }finally{busy=false;updateSendState();}
}
function download(name,contents,type){
 const blob=new Blob([contents],{type});
 const url=URL.createObjectURL(blob);
 const a=document.createElement("a");
 a.href=url;a.download=name;document.body.append(a);a.click();a.remove();
 setTimeout(()=>URL.revokeObjectURL(url),5000);
}
function setExample(){
 $("url").value="https://petstore3.swagger.io/api/v3/openapi.json";
 $("approved-host").value="petstore3.swagger.io";
 $("method").value="GET";
 $("timeout").value="10000";
 $("headers").value='{"Accept":"application/json"}';
 $("permission").checked=false;$("credentials").checked=false;
 last=null;response.hidden=true;updateSendState();
 setStatus("Example configured but NOT sent. Review site terms, confirm authorization, then press Send.");
}
function reset(){
 $("url").value="";$("approved-host").value="";$("method").value="GET";
 $("headers").value='{"Accept":"application/json"}';
 $("permission").checked=false;$("credentials").checked=false;
 last=null;response.hidden=true;
 updateSendState();setStatus("Cleared. No request made or stored.");
}
for(const id of ["url","approved-host","method","timeout","headers","permission","credentials"]){
 const field=$(id);field.addEventListener("input",updateSendState);field.addEventListener("change",updateSendState);
}
$("send").addEventListener("click",send);
$("example").addEventListener("click",setExample);
$("reset").addEventListener("click",reset);
$("copy-response").addEventListener("click",async()=>{
 if(!last)return;
 try{await navigator.clipboard.writeText(last.responseBody);setStatus("Response body copied to clipboard.","success");}
 catch{setStatus("Clipboard unavailable. Select and copy text from the response box.","error");}
});
$("download-report").addEventListener("click",()=>{
 if(!last)return;
 download("restless42-request-evidence.json",JSON.stringify(last,null,2),"application/json;charset=utf-8");
 setStatus("Response evidence saved to a local JSON file. Request authorization headers are not included.","success");
});
updateSendState();
setStatus("Ready. This page does not access the network without explicit permission and a Send click.");
