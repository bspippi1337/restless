// Real Chromium smoke test for the generated RESTLESS GitHub Pages website.
// Run from repo root with PLAYWRIGHT_PACKAGE pointing to an installed playwright/index.mjs.
import {createServer} from "node:http";
import {readFile,stat,mkdir} from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import {pathToFileURL} from "node:url";
import assert from "node:assert/strict";

const packageFile=process.env.PLAYWRIGHT_PACKAGE;
if(!packageFile)throw Error("Set PLAYWRIGHT_PACKAGE to the installed playwright/index.mjs.");
const {chromium}=await import(pathToFileURL(path.resolve(packageFile)).href);
const site=path.resolve(process.argv[2]||"site");
const screens=path.resolve(process.argv[3]||path.join(os.tmpdir(),"restless-browser-preview"));
await mkdir(screens,{recursive:true});
const mime={
 ".html":"text/html; charset=utf-8",".css":"text/css; charset=utf-8",
 ".js":"text/javascript; charset=utf-8",".mjs":"text/javascript; charset=utf-8",
 ".json":"application/json; charset=utf-8",".svg":"image/svg+xml",
 ".txt":"text/plain; charset=utf-8",".md":"text/markdown; charset=utf-8"
};
const server=createServer(async(req,res)=>{
  try{
    const url=new URL(req.url,"http://127.0.0.1");
    if(!url.pathname.startsWith("/restless/")){res.writeHead(404);res.end();return;}
    const pathname=decodeURIComponent(url.pathname.slice("/restless/".length));
    const safe=path.resolve(site,pathname);
    if(safe!==site&&!safe.startsWith(site+path.sep)){res.writeHead(403);res.end();return;}
    let file=safe;
    const f=await stat(file);
    if(f.isDirectory())file=path.join(file,"index.html");
    const data=await readFile(file);
    res.writeHead(200,{"Content-Type":mime[path.extname(file)]||"application/octet-stream","X-Content-Type-Options":"nosniff"});
    res.end(data);
  }catch{res.writeHead(404);res.end("Not found");}
});
await new Promise(resolve=>server.listen(0,"127.0.0.1",resolve));
const base="http://127.0.0.1:"+server.address().port+"/restless/";
let browser;
try{
  browser=await chromium.launch({channel:"chrome",headless:true,args:["--no-sandbox"]});
  const desktop=await browser.newPage({viewport:{width:1440,height:960}});
  const errors=[],externalRequests=[];
  desktop.on("pageerror",e=>errors.push(e.message));
  desktop.on("request",req=>{if(!req.url().startsWith(base))externalRequests.push(req.url());});
  const response=await desktop.goto(base,{waitUntil:"domcontentloaded",timeout:30000});
  assert.equal(response.status(),200);
  await desktop.locator(".operation-item").first().waitFor({timeout:15000});
  assert.equal(await desktop.locator(".operation-item").count(),32,"32 operations must be visible on homepage");
  assert.equal(await desktop.locator("#preset").count(),1,"the root must serve the REAL toolbox interface");
  await desktop.selectOption("#preset","json");
  await desktop.waitForFunction(()=>document.querySelector("#output")?.value.includes('"project": "RESTLESS 42"'));
  assert.match(await desktop.locator("#status").innerText(),/Complete/);
  await desktop.locator("#clear-recipe").click();
  await desktop.locator("#input").fill("abc");
  await desktop.locator("#operation-search").fill("SHA-256");
  assert.equal(await desktop.locator(".operation-item").count(),1);
  await desktop.locator(".operation-item").first().click();
  await desktop.waitForFunction(()=>document.querySelector("#output")?.value.startsWith("ba7816bf8f01cfea"));
  await desktop.locator("#operation-search").fill("");
  await desktop.selectOption("#preset","openapi");
  await desktop.waitForFunction(()=>document.querySelector("#output")?.value.includes('"DOCUMENTATION_ONLY_NO_EXECUTION"'));
  assert.equal(errors.length,0,"Browser runtime errors: "+errors.join(" | "));
  assert.equal(externalRequests.length,0,"Toolbox must not make external network requests.");
  await desktop.screenshot({path:path.join(screens,"RESTLESS-42-Toolbox-desktop.png"),fullPage:true});

  const mobile=await browser.newPage({viewport:{width:390,height:844},deviceScaleFactor:1});
  const mobileErrors=[];
  mobile.on("pageerror",e=>mobileErrors.push(e.message));
  await mobile.goto(base,{waitUntil:"domcontentloaded"});
  await mobile.locator(".operation-item").first().waitFor();
  assert.equal(await mobile.locator(".operation-item").count(),32);
  await mobile.selectOption("#preset","encode");
  await mobile.waitForFunction(()=>document.querySelector("#output")?.value.length>10);
  assert.equal(mobileErrors.length,0,"Mobile runtime errors: "+mobileErrors.join(" | "));
  await mobile.screenshot({path:path.join(screens,"RESTLESS-42-Toolbox-mobile.png"),fullPage:true});

  const guide=await desktop.goto(base+"guide/",{waitUntil:"domcontentloaded"});
  assert.equal(guide.status(),200);
  assert.ok(await desktop.locator("article h1,article h2").count()>=1);
  const contract=await desktop.goto(base+"contract/",{waitUntil:"domcontentloaded"});
  assert.equal(contract.status(),200);
  await desktop.locator("#example").click();
  await desktop.locator("#results").waitFor({state:"visible"});
  assert.match(await desktop.locator("#contract-status").innerText(),/zero network requests/i);
  assert.equal(errors.length,0,"Site JS errors: "+errors.join(" | "));

  console.log("PASS: homepage Toolbox, 32 operations, JSON, SHA-256, OpenAPI, mobile UI, docs, Contract Lab.");
  console.log("PASS: no unexpected browser exceptions or external API requests.");
  console.log("Screenshots:",screens);
}finally{await browser?.close();await new Promise(resolve=>server.close(resolve));}
