import {OPERATIONS,OPERATION_BY_ID,validateRecipe,runRecipe} from "./transforms.mjs";
const $=id=>document.getElementById(id);
const ui={search:$("operation-search"),operations:$("operation-list"),operationCount:$("operation-count"),recipe:$("recipe-list"),recipeCount:$("recipe-count"),input:$("input"),output:$("output"),inputMeta:$("input-meta"),outputMeta:$("output-meta"),timing:$("timing"),status:$("status"),auto:$("autobake"),preset:$("preset")};
const encoder=new TextEncoder();
let recipe=[],timer=null,generation=0;
function element(tag,className,text){const el=document.createElement(tag);if(className)el.className=className;if(text!==undefined)el.textContent=text;return el;}
function setStatus(message,kind="info"){ui.status.textContent=message;ui.status.dataset.kind=kind;}
function refreshInputMeta(){ui.inputMeta.textContent=[...ui.input.value].length.toLocaleString()+" characters · "+encoder.encode(ui.input.value).length.toLocaleString()+" UTF-8 bytes";}
function refreshOutputMeta(steps=0){ui.outputMeta.textContent=encoder.encode(ui.output.value).length.toLocaleString()+" UTF-8 bytes · "+steps+" executed steps";}
function queueBake(){generation++;if(timer)clearTimeout(timer);if(ui.auto.checked)timer=setTimeout(()=>bake(),230);}
async function bake(){
  if(timer)clearTimeout(timer);timer=null;
  const run=++generation,started=performance.now();ui.outputMeta.textContent="Processing...";
  ui.recipe.querySelectorAll("[data-error]").forEach(el=>delete el.dataset.error);
  try{const result=await runRecipe(ui.input.value,recipe);if(run!==generation)return;ui.output.value=result.output;refreshOutputMeta(result.history.length);ui.timing.textContent=Math.round(performance.now()-started)+" ms";setStatus("Complete. "+result.history.length+" operation(s) executed locally. No server request.", "success");}
  catch(error){if(run!==generation)return;ui.output.value="";ui.outputMeta.textContent="Operation failed";ui.timing.textContent="";setStatus(error.message,"error");if(Number.isInteger(error.step)){const item=ui.recipe.querySelector('[data-index="'+error.step+'"]');if(item)item.dataset.error="true";}}
}
function renderPalette(){
  const query=ui.search.value.trim().toLowerCase();
  const shown=OPERATIONS.filter(op=>[op.label,op.group,op.description].join(" ").toLowerCase().includes(query));
  ui.operations.replaceChildren();ui.operationCount.textContent=shown.length+" / "+OPERATIONS.length;
  let group="";
  for(const op of shown){
    if(op.group!==group){group=op.group;ui.operations.append(element("h3","operation-group",group));}
    const button=element("button","operation-item");button.type="button";button.title=op.description;button.setAttribute("aria-label","Add "+op.label+" to recipe");
    button.append(element("span","",op.label),element("span","add-icon","+"));
    button.addEventListener("click",()=>{if(recipe.length>=24){setStatus("Recipe limit: 24 operations.","error");return;}recipe.push({id:op.id,enabled:true,option:op.defaultOption||""});renderRecipe();queueBake();setStatus("Added "+op.label+".");});
    ui.operations.append(button);
  }
  if(!shown.length)ui.operations.append(element("p","operation-empty","No matching operations."));
}
function stepButton(text,title,onClick,disabled=false){
  const b=element("button","",text);b.type="button";b.title=title;b.setAttribute("aria-label",title);b.disabled=disabled;b.addEventListener("click",onClick);return b;
}
function renderRecipe(){
  ui.recipe.replaceChildren();ui.recipeCount.textContent=recipe.length+" step"+(recipe.length===1?"":"s");
  if(!recipe.length){ui.recipe.append(element("li","recipe-empty","Your recipe is empty. Click an operation on the left to start."));return;}
  recipe.forEach((item,i)=>{
    const op=OPERATION_BY_ID.get(item.id);
    const li=element("li","recipe-step");li.dataset.index=String(i);if(!item.enabled)li.dataset.disabled="true";
    const main=element("div","step-main");
    const name=element("span","step-title",op.label),controls=element("div","step-tools");
    const swap=(to)=>{const tmp=recipe[i];recipe[i]=recipe[to];recipe[to]=tmp;renderRecipe();queueBake();};
    controls.append(
      stepButton("↑","Move "+op.label+" up",()=>swap(i-1),i===0),
      stepButton("↓","Move "+op.label+" down",()=>swap(i+1),i===recipe.length-1),
      stepButton(item.enabled?"◉":"○",item.enabled?"Disable "+op.label:"Enable "+op.label,()=>{item.enabled=!item.enabled;renderRecipe();queueBake();}),
      stepButton("×","Remove "+op.label,()=>{recipe.splice(i,1);renderRecipe();queueBake();})
    );
    main.append(element("span","step-index",String(i+1).padStart(2,"0")),name,controls);
    li.append(main,element("p","step-description",op.description));
    if(op.option){
      const label=element("label","step-option");
      label.append(element("span","",op.option));
      const input=element("input");input.type="text";input.value=item.option||"";input.maxLength=500;input.setAttribute("aria-label",op.option+" for "+op.label);
      input.addEventListener("input",()=>{item.option=input.value;queueBake();});
      label.append(input);li.append(label);
    }
    ui.recipe.append(li);
  });
}
function download(name,content,type="text/plain;charset=utf-8"){
  const url=URL.createObjectURL(new Blob([content],{type}));
  const a=document.createElement("a");a.href=url;a.download=name;document.body.append(a);a.click();a.remove();setTimeout(()=>URL.revokeObjectURL(url),5000);
}
async function loadInput(file){
  if(!file)return;
  if(file.size>2_000_000){setStatus("Files must be 2 MB or smaller. Nothing was loaded.","error");return;}
  try{const text=await file.text();if(text.length>2_000_000)throw Error("Text exceeds 2 million characters.");ui.input.value=text;refreshInputMeta();queueBake();setStatus('Loaded "'+file.name+'" locally.');}
  catch(e){setStatus("Cannot load file: "+e.message,"error");}
}
function preset(name){
  switch(name){
    case"json":recipe=[{id:"json-sort",enabled:true,option:""}];ui.input.value='{"project":"RESTLESS 42","active":true,"operations":["harvest","index","plan"],"metadata":{"license":"MIT","local":true}}';break;
    case"encode":recipe=[{id:"base64-encode",enabled:true,option:""},{id:"hex-encode",enabled:true,option:""}];ui.input.value="Quiet API Discovery · RESTLESS 42";break;
    case"jwt":recipe=[{id:"jwt-inspect",enabled:true,option:""}];ui.input.value="eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJkZW1vIiwiaXNzIjoicmVzdGxlc3MtNDIifQ.";break;
    case"openapi":recipe=[{id:"openapi-plan",enabled:true,option:"find pet by id"}];ui.input.value=JSON.stringify({openapi:"3.0.3",info:{title:"RESTLESS demo Pet API",version:"1"},paths:{"/pets":{get:{summary:"List pets",operationId:"listPets",tags:["pets"]},post:{summary:"Create a pet",operationId:"createPet",tags:["pets"],security:[{api_key:[]}]}}, "/pets/{petId}":{get:{summary:"Find pet by ID",operationId:"findPetById",parameters:[{name:"petId",in:"path",required:true,schema:{type:"string"}}]}}}},null,2);break;
    case"csv":recipe=[{id:"csv-to-json",enabled:true,option:""}];ui.input.value='name,role\n"Alex, Jr.",Engineer\nSam,Researcher';break;
    case"digest":recipe=[{id:"sha-256",enabled:true,option:""}];ui.input.value="RESTLESS 42";break;
    default:return;
  }
  renderRecipe();refreshInputMeta();queueBake();setStatus("Loaded example. Everything happens in this browser.");
}
function bind(){
  ui.search.addEventListener("input",renderPalette);
  ui.input.addEventListener("input",()=>{refreshInputMeta();queueBake();});
  ui.auto.addEventListener("change",()=>{if(ui.auto.checked)queueBake();else{if(timer)clearTimeout(timer);timer=null;setStatus("Auto-bake disabled. Use Bake recipe to process.");}});
  $("bake").addEventListener("click",bake);
  $("clear-input").addEventListener("click",()=>{ui.input.value="";refreshInputMeta();queueBake();});
  $("clear-recipe").addEventListener("click",()=>{recipe=[];renderRecipe();queueBake();setStatus("Recipe cleared.");});
  $("output-to-input").addEventListener("click",()=>{if(!ui.output.value){setStatus("No output to reuse.","error");return;}ui.input.value=ui.output.value;refreshInputMeta();queueBake();setStatus("Previous output moved into input.");});
  $("save-output").addEventListener("click",()=>{download("restless-42-output.txt",ui.output.value);setStatus("Output downloaded as a local text file.");});
  $("copy-output").addEventListener("click",async()=>{
    try{if(navigator.clipboard?.writeText)await navigator.clipboard.writeText(ui.output.value);else{ui.output.focus();ui.output.select();if(!document.execCommand("copy"))throw Error("Clipboard API unavailable");}setStatus("Output copied to clipboard.","success");}
    catch(e){setStatus("Clipboard blocked. Select and copy the output manually.","error");}
  });
  $("export-recipe").addEventListener("click",()=>{download("restless-42-recipe.json",JSON.stringify({format:"restless-42-toolbox-recipe",version:1,steps:recipe},null,2),"application/json");setStatus("Recipe exported. The input text is not included.");});
  $("import-recipe").addEventListener("click",()=>$("recipe-file").click());
  $("recipe-file").addEventListener("change",async e=>{const file=e.target.files?.[0];e.target.value="";if(!file)return;if(file.size>50_000){setStatus("Recipe JSON is too large.","error");return;}try{const obj=JSON.parse(await file.text());if(obj.format!=="restless-42-toolbox-recipe"||obj.version!==1)throw Error("Unsupported recipe format/version.");recipe=validateRecipe(obj.steps);renderRecipe();queueBake();setStatus("Imported "+recipe.length+" operations; input text unchanged.","success");}catch(err){setStatus("Cannot import recipe: "+err.message,"error");}});
  ui.preset.addEventListener("change",()=>{preset(ui.preset.value);ui.preset.value="";});
  $("load-file").addEventListener("click",()=>$("input-file").click());
  $("input-file").addEventListener("change",async e=>{const file=e.target.files?.[0];e.target.value="";await loadInput(file);});
  ui.input.addEventListener("dragover",e=>{if(e.dataTransfer?.types.includes("Files"))e.preventDefault();});
  ui.input.addEventListener("drop",async e=>{const file=e.dataTransfer?.files?.[0];if(file){e.preventDefault();await loadInput(file);}});
  document.addEventListener("keydown",e=>{if((e.ctrlKey||e.metaKey)&&e.key==="Enter"){e.preventDefault();bake();}});
}
bind();renderPalette();renderRecipe();refreshInputMeta();refreshOutputMeta();
setStatus("Ready. Choose an example or add operations. Data stays in your browser.");
