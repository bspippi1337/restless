// Single source of truth for both GitHub Pages and release-archive documentation.
import {copyFile,mkdir} from "node:fs/promises";
import {execFileSync} from "node:child_process";
import path from "node:path";
const root=path.resolve(process.argv[2]||".");
const dest=path.resolve(root,"site");
async function pandoc(source,output,template,title){
 await mkdir(path.dirname(output),{recursive:true});
 execFileSync("pandoc",[source,"--from=gfm","--to=html5","--standalone",
   "--toc","--toc-depth=2","--metadata","title="+title,
   "--template="+template,"--output="+output],{cwd:root,stdio:"inherit"});
}
await mkdir(path.join(dest,"downloads"),{recursive:true});
await pandoc("docs/RESTLESS_42_Field_Guide_EN.md","site/guide/index.html",
  "tools/pages/guide-template.html","The Field Guide to Quiet API Discovery");
await pandoc("docs/RESTLESS_42_Browser_Toolbox_EN.md","site/toolbox/manual/index.html",
  "tools/pages/toolbox-template.html","Browser Toolbox Manual");
const bundle=[
  "RESTLESS_42_Field_Guide_EN.md",
  "RESTLESS_42_Browser_Toolbox_EN.md",
  "RESTLESS_42_NotebookLM_Podcast_Brief_EN.txt"
];
if(await import("node:fs").then(m=>m.existsSync(path.join(root,"docs/RESTLESS_42_Request_Lab_EN.md")))){
 await pandoc("docs/RESTLESS_42_Request_Lab_EN.md","site/request/manual/index.html",
   "tools/pages/request-template.html","Request Lab Manual");
 bundle.push("RESTLESS_42_Request_Lab_EN.md");
}
for(const file of bundle)await copyFile(path.join(root,"docs",file),path.join(dest,"downloads",file));
console.log("RESTLESS Pages docs rendered: "+bundle.join(", "));
