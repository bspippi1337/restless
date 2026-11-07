// Check built GitHub Pages site without network or dependencies.
import {readdir,readFile,stat} from "node:fs/promises";
import path from "node:path";
import {pathToFileURL,fileURLToPath} from "node:url";
const root=path.resolve(process.argv[2]||"site");
const failures=[],pages=[];
async function visit(directory){
  for(const item of await readdir(directory,{withFileTypes:true})){
    const full=path.join(directory,item.name);
    if(item.isDirectory())await visit(full);
    else if(item.isFile()&&item.name.endsWith(".html"))pages.push(full);
  }
}
await visit(root);
if(!pages.length)failures.push("No rendered HTML pages found.");
for(const page of pages){
  const text=await readFile(page,"utf8");
  if(!/<main\b[^>]*>/i.test(text))failures.push(path.relative(root,page)+": missing <main>");
  if(!/<title>[^<]+<\/title>/i.test(text))failures.push(path.relative(root,page)+": missing title");
  const attributes=/<(?:a|link|script|img)\b[^>]*?\b(?:href|src)\s*=\s*["']([^"']+)["']/gi;
  for(const match of text.matchAll(attributes)){
    const uri=match[1];
    if(/^(https?:|mailto:|tel:|data:|#)/i.test(uri))continue;
    try{
      const url=new URL(uri,pathToFileURL(page));
      if(url.protocol!=="file:")continue;
      let p=fileURLToPath(url);
      const info=await stat(p);
      if(info.isDirectory()){p=path.join(p,"index.html");await stat(p);}
      if(!path.resolve(p).startsWith(root+path.sep))failures.push(path.relative(root,page)+": link escapes site: "+uri);
    }catch(e){failures.push(path.relative(root,page)+": broken local asset or link "+uri);}
  }
}
for(const required of ["index.html","guide/index.html","podcast/index.html","toolbox/index.html","toolbox/manual/index.html","docs/index.html","toolbox/app.mjs","toolbox/transforms.mjs","downloads/RESTLESS_42_Field_Guide_EN.md","downloads/RESTLESS_42_Browser_Toolbox_EN.md","downloads/RESTLESS_42_NotebookLM_Podcast_Brief_EN.txt"]){
  try{const info=await stat(path.join(root,required));if(!info.isFile())throw Error("not file");}
  catch{failures.push("Missing site asset: "+required);}
}
if(failures.length){console.error("RESTLESS Pages validation failed:\n"+failures.join("\n"));process.exit(1);}
console.log("RESTLESS Pages validation passed: "+pages.length+" HTML pages and all local links/assets resolved.");
