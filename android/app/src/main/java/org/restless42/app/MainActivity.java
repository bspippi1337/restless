package org.restless42.app;

import android.app.Activity;
import android.app.AlertDialog;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.content.Intent;
import android.graphics.Color;
import android.graphics.Typeface;
import android.graphics.drawable.GradientDrawable;
import android.net.Uri;
import android.os.Build;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.text.Editable;
import android.text.InputType;
import android.text.TextWatcher;
import android.view.Gravity;
import android.view.View;
import android.view.ViewGroup;
import android.view.WindowInsets;
import android.widget.AdapterView;
import android.widget.ArrayAdapter;
import android.widget.Button;
import android.widget.CheckBox;
import android.widget.EditText;
import android.widget.FrameLayout;
import android.widget.LinearLayout;
import android.widget.ScrollView;
import android.widget.Spinner;
import android.widget.TextView;
import android.widget.Toast;

import org.json.JSONObject;

import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import java.io.OutputStream;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/** Native RESTLESS 42 UI. No WebView and no network permission. */
public final class MainActivity extends Activity {
    private static final int BG=Color.rgb(11,14,21),SURFACE=Color.rgb(22,29,41),
        SURFACE_2=Color.rgb(31,40,54),STROKE=Color.rgb(60,76,96),
        WHITE=Color.rgb(238,246,255),MUTED=Color.rgb(172,188,207),
        GREEN=Color.rgb(183,247,104),RED=Color.rgb(255,157,153);
    private static final int READ_INPUT=101,IMPORT_RECIPE=102,SAVE_OUTPUT=103,SAVE_RECIPE=104;
    private final List<ToolboxEngine.Step> recipe=new ArrayList<>();
    private final Handler main=new Handler(Looper.getMainLooper());
    private final ExecutorService workers=Executors.newSingleThreadExecutor();
    private int generation=0;
    private FrameLayout content;
    private LinearLayout stepsContainer;
    private EditText sourceText,outputText;
    private TextView statusText,recipeCount;
    private Spinner toolSpinner;
    private CheckBox autoRun;
    private Runnable pending;
    private int selectedTab=0;
    private String currentOutput="";

    @Override public void onCreate(Bundle state){
        super.onCreate(state);
        getWindow().setStatusBarColor(BG);
        getWindow().setNavigationBarColor(BG);
        getWindow().getDecorView().setSystemUiVisibility(0);
        createUI();
    }
    @Override protected void onDestroy(){generation++;workers.shutdownNow();super.onDestroy();}
    private int dp(float value){return (int)(getResources().getDisplayMetrics().density*value+.5f);}
    private GradientDrawable panel(int color,int radius,int line){
        GradientDrawable bg=new GradientDrawable();
        bg.setColor(color);bg.setCornerRadius(dp(radius));
        if(line!=0)bg.setStroke(dp(1),line);
        return bg;
    }
    private LinearLayout vertical(){LinearLayout l=new LinearLayout(this);l.setOrientation(LinearLayout.VERTICAL);return l;}
    private LinearLayout horizontal(){LinearLayout l=new LinearLayout(this);l.setOrientation(LinearLayout.HORIZONTAL);l.setGravity(Gravity.CENTER_VERTICAL);return l;}
    private TextView text(String value,int size,int color,boolean bold){
        TextView t=new TextView(this);t.setText(value);t.setTextSize(size);t.setTextColor(color);
        if(bold)t.setTypeface(Typeface.DEFAULT,Typeface.BOLD);
        t.setGravity(Gravity.CENTER_VERTICAL);
        return t;
    }
    private Button action(String label,boolean primary,Runnable press){
        Button b=new Button(this);b.setText(label);b.setTextSize(12);b.setAllCaps(false);
        b.setTypeface(Typeface.DEFAULT,Typeface.BOLD);
        b.setTextColor(primary?BG:WHITE);
        b.setBackground(panel(primary?GREEN:SURFACE_2,9,primary?0:STROKE));
        b.setPadding(dp(12),0,dp(12),0);
        b.setMinHeight(dp(43));b.setMinimumHeight(dp(43));
        b.setOnClickListener(v->press.run());
        return b;
    }
    private LinearLayout.LayoutParams lp(int width,int height,int bottom){
        LinearLayout.LayoutParams p=new LinearLayout.LayoutParams(width<0?width:dp(width),height<0?height:dp(height));
        p.bottomMargin=dp(bottom);return p;
    }
    private void addButton(LinearLayout target,Button button){
        LinearLayout.LayoutParams p=lp(-2,44,0);p.rightMargin=dp(7);target.addView(button,p);
    }
    private void createUI(){
        LinearLayout root=vertical();root.setBackgroundColor(BG);
        if(Build.VERSION.SDK_INT>=35){
            root.setOnApplyWindowInsetsListener((v,insets)->{
                android.graphics.Insets bars=insets.getInsets(WindowInsets.Type.systemBars());
                v.setPadding(0,bars.top,0,bars.bottom);return insets;
            });
        }
        setContentView(root);
        LinearLayout bar=horizontal();bar.setBackgroundColor(BG);bar.setPadding(dp(18),dp(14),dp(18),dp(10));
        TextView emblem=text("R.",27,GREEN,true);emblem.setContentDescription("RESTLESS");bar.addView(emblem,lp(43,40,0));
        LinearLayout name=vertical();name.addView(text("RESTLESS 42",19,WHITE,true));name.addView(text("NATIVE  /  OFFLINE",10,GREEN,true));
        bar.addView(name,new LinearLayout.LayoutParams(0,-2,1f));
        TextView badge=text("NO NETWORK",10,GREEN,true);badge.setPadding(dp(10),dp(7),dp(10),dp(7));badge.setBackground(panel(SURFACE_2,8,STROKE));
        bar.addView(badge);root.addView(bar,lp(-1,72,0));

        LinearLayout nav=horizontal();nav.setPadding(dp(12),dp(5),dp(8),dp(7));
        Button work=action("Workbench",true,()->showTab(0)),guide=action("Guide",false,()->showTab(1)),about=action("About",false,()->showTab(2));
        addButton(nav,work);addButton(nav,guide);addButton(nav,about);
        root.addView(nav,lp(-1,54,0));
        View line=new View(this);line.setBackgroundColor(STROKE);root.addView(line,lp(-1,1,0));
        content=new FrameLayout(this);root.addView(content,new LinearLayout.LayoutParams(-1,0,1));
        showTab(0);
    }
    private void showTab(int tab){
        selectedTab=tab;
        content.removeAllViews();
        if(tab==0)showWorkbench();
        if(tab==1)showGuide();
        if(tab==2)showAbout();
    }
    private ScrollView scroller(LinearLayout holder){
        ScrollView scroll=new ScrollView(this);scroll.setFillViewport(false);scroll.setClipToPadding(false);
        scroll.addView(holder,new ScrollView.LayoutParams(-1,-2));return scroll;
    }
    private LinearLayout paddedContent(){
        LinearLayout main=vertical();main.setPadding(dp(16),dp(23),dp(16),dp(35));return main;
    }
    private void heading(LinearLayout root,String title,String subtitle){
        root.addView(text(title,25,WHITE,true),lp(-1,-2,4));
        TextView t=text(subtitle,13,MUTED,false);
        t.setLineSpacing(dp(2),1);root.addView(t,lp(-1,-2,18));
    }
    private EditText editor(String hint,int minLines){
        EditText ed=new EditText(this);
        ed.setTextColor(WHITE);ed.setHintTextColor(MUTED);ed.setTextSize(13);
        ed.setTypeface(Typeface.MONOSPACE);
        ed.setGravity(Gravity.TOP|Gravity.START);
        ed.setPadding(dp(15),dp(12),dp(15),dp(12));
        ed.setMinLines(minLines);ed.setMaxLines(18);
        ed.setInputType(InputType.TYPE_CLASS_TEXT|InputType.TYPE_TEXT_FLAG_MULTI_LINE|InputType.TYPE_TEXT_FLAG_NO_SUGGESTIONS);
        ed.setHorizontalScrollBarEnabled(true);
        ed.setBackground(panel(SURFACE,11,STROKE));ed.setHint(hint);return ed;
    }
    private void subsection(LinearLayout main,String label){
        TextView t=text(label.toUpperCase(),11,GREEN,true);t.setLetterSpacing(.12f);main.addView(t,lp(-1,-2,9));
    }
    private void showWorkbench(){
        LinearLayout main=paddedContent();
        heading(main,"An entire API workbench.","32 native text and API-document operations. Combine them into a recipe. Nothing is sent to a server.");
        LinearLayout sample=horizontal();
        addButton(sample,action("JSON sample",false,()->preset("json")));
        addButton(sample,action("OpenAPI sample",false,()->preset("openapi")));
        addButton(sample,action("JWT sample",false,()->preset("jwt")));
        main.addView(sample,lp(-1,-2,17));

        subsection(main,"Input");
        sourceText=editor("Paste JSON, encoded text, or an OpenAPI specification...",7);
        main.addView(sourceText,lp(-1,-2,7));
        sourceText.addTextChangedListener(new TextWatcher(){
            public void beforeTextChanged(CharSequence s,int st,int c,int a){}
            public void onTextChanged(CharSequence s,int st,int before,int count){queueBake();}
            public void afterTextChanged(Editable e){}
        });
        LinearLayout inputActions=horizontal();
        addButton(inputActions,action("Open file",false,()->openText(READ_INPUT,"*/*")));
        addButton(inputActions,action("Paste",false,()->paste()));
        addButton(inputActions,action("Clear",false,()->sourceText.setText("")));
        main.addView(inputActions,lp(-1,-2,20));

        subsection(main,"Recipe");
        LinearLayout tools=horizontal();
        toolSpinner=new Spinner(this);
        List<String> labels=new ArrayList<>();for(ToolboxEngine.Tool tool:ToolboxEngine.TOOLS)labels.add(tool.category+"  /  "+tool.label);
        ArrayAdapter<String> adapter=new ArrayAdapter<>(this,android.R.layout.simple_spinner_item,labels);
        adapter.setDropDownViewResource(android.R.layout.simple_spinner_dropdown_item);
        toolSpinner.setAdapter(adapter);
        toolSpinner.setBackground(panel(SURFACE_2,8,STROKE));
        tools.addView(toolSpinner,new LinearLayout.LayoutParams(0,dp(50),1));
        Button add=action("+ Add",true,this::addTool);
        LinearLayout.LayoutParams addParams=lp(86,50,0);addParams.leftMargin=dp(7);
        tools.addView(add,addParams);
        main.addView(tools,lp(-1,52,8));

        recipeCount=text("0 steps",12,MUTED,false);main.addView(recipeCount,lp(-1,-2,5));
        stepsContainer=vertical();main.addView(stepsContainer,lp(-1,-2,10));
        LinearLayout recipeActions=horizontal();
        addButton(recipeActions,action("Import",false,()->openText(IMPORT_RECIPE,"application/json")));
        addButton(recipeActions,action("Export",false,()->saveText(SAVE_RECIPE,"restless42-recipe.json")));
        addButton(recipeActions,action("Reset",false,()->{recipe.clear();renderSteps();queueBake();}));
        main.addView(recipeActions,lp(-1,-2,12));

        LinearLayout runLine=horizontal();
        autoRun=new CheckBox(this);autoRun.setText("Auto run");autoRun.setTextColor(MUTED);autoRun.setButtonTintList(android.content.res.ColorStateList.valueOf(GREEN));
        autoRun.setChecked(false);autoRun.setOnCheckedChangeListener((button,isChecked)->{if(isChecked)queueBake();});
        runLine.addView(autoRun,new LinearLayout.LayoutParams(0,-2,1));
        addButton(runLine,action("▶ RUN RECIPE",true,this::bake));
        main.addView(runLine,lp(-1,52,15));

        subsection(main,"Output");
        outputText=editor("The native engine's output appears here.",8);
        outputText.setFocusable(false);outputText.setTextIsSelectable(true);outputText.setKeyListener(null);
        main.addView(outputText,lp(-1,-2,7));
        LinearLayout outputActions=horizontal();
        addButton(outputActions,action("Copy",false,this::copy));
        addButton(outputActions,action("Save .txt",false,()->saveText(SAVE_OUTPUT,"restless42-output.txt")));
        addButton(outputActions,action("Use as input",false,()->sourceText.setText(currentOutput)));
        main.addView(outputActions,lp(-1,-2,13));
        statusText=text("Ready. Add an operation and press Run recipe.",12,MUTED,false);
        statusText.setBackground(panel(SURFACE_2,10,STROKE));statusText.setPadding(dp(13),dp(12),dp(13),dp(12));
        main.addView(statusText,lp(-1,-2,12));
        main.addView(text("LOCAL BY DESIGN   /   NO WEBVIEW   /   NO INTERNET PERMISSION",10,GREEN,true),lp(-1,-2,0));
        content.addView(scroller(main));
        renderSteps();
    }
    private void addTool(){
        if(recipe.size()>=24){notice("Maximum 24 steps in one recipe.",true);return;}
        ToolboxEngine.Tool t=ToolboxEngine.TOOLS.get(toolSpinner.getSelectedItemPosition());
        recipe.add(new ToolboxEngine.Step(t.id,t.defaultOption,true));renderSteps();queueBake();
    }
    private void renderSteps(){
        if(stepsContainer==null)return;
        stepsContainer.removeAllViews();
        recipeCount.setText(recipe.size()+" step"+(recipe.size()==1?"":"s"));
        if(recipe.isEmpty()){
            TextView empty=text("No operations yet. Choose an operation above or load an example.",13,MUTED,false);
            empty.setPadding(dp(14),dp(20),dp(14),dp(20));empty.setBackground(panel(SURFACE,10,STROKE));
            stepsContainer.addView(empty,lp(-1,-2,9));return;
        }
        for(int index=0;index<recipe.size();index++){
            final int i=index;
            ToolboxEngine.Step step=recipe.get(i);ToolboxEngine.Tool tool=ToolboxEngine.find(step.id);
            LinearLayout card=vertical();card.setPadding(dp(9),dp(8),dp(9),dp(10));
            card.setBackground(panel(SURFACE,9,STROKE));
            LinearLayout row=horizontal();
            CheckBox enabled=new CheckBox(this);enabled.setChecked(step.enabled);enabled.setButtonTintList(android.content.res.ColorStateList.valueOf(GREEN));
            enabled.setContentDescription("Enable "+tool.label);
            enabled.setOnCheckedChangeListener((v,value)->{step.enabled=value;queueBake();});
            row.addView(enabled,lp(39,42,0));
            TextView label=text((i+1)+". "+tool.label,13,WHITE,true);
            row.addView(label,new LinearLayout.LayoutParams(0,-2,1));
            addButton(row,action("↑",false,()->move(i,-1)));
            addButton(row,action("↓",false,()->move(i,1)));
            addButton(row,action("×",false,()->{recipe.remove(i);renderSteps();queueBake();}));
            card.addView(row,lp(-1,-2,0));
            if(!tool.optionLabel.isEmpty()){
                TextView prompt=text(tool.optionLabel,11,MUTED,false);card.addView(prompt,lp(-1,-2,4));
                EditText option=editor(tool.defaultOption,1);option.setSingleLine(true);
                option.setText(step.option);option.setTextSize(13);
                option.addTextChangedListener(new TextWatcher(){
                    public void beforeTextChanged(CharSequence s,int start,int count,int after){}
                    public void onTextChanged(CharSequence s,int start,int before,int count){step.option=s.toString();queueBake();}
                    public void afterTextChanged(Editable s){}
                });
                card.addView(option,lp(-1,42,0));
            }
            stepsContainer.addView(card,lp(-1,-2,8));
        }
    }
    private void move(int index,int delta){
        int next=index+delta;if(next<0||next>=recipe.size())return;
        java.util.Collections.swap(recipe,index,next);renderSteps();queueBake();
    }
    private void queueBake(){
        generation++;
        if(pending!=null)main.removeCallbacks(pending);
        if(autoRun!=null&&autoRun.isChecked()&&sourceText!=null&&outputText!=null){
            pending=this::bake;main.postDelayed(pending,320);
        }
    }
    private void bake(){
        if(sourceText==null||outputText==null)return;
        if(pending!=null)main.removeCallbacks(pending);
        final int ticket=++generation;
        final String source=sourceText.getText().toString();
        final List<ToolboxEngine.Step> steps=new ArrayList<>();
        for(ToolboxEngine.Step step:recipe)steps.add(new ToolboxEngine.Step(step.id,step.option,step.enabled));
        notice("Processing "+steps.size()+" step(s) on this device...",false);
        workers.execute(()->{
            try{
                String result=ToolboxEngine.run(source,steps);
                main.post(()->{
                    if(ticket!=generation||selectedTab!=0)return;
                    currentOutput=result;outputText.setText(result);
                    notice("Success. "+steps.size()+" native steps. No network access.",false);
                });
            }catch(Exception e){
                main.post(()->{
                    if(ticket!=generation||selectedTab!=0)return;
                    currentOutput="";outputText.setText("");
                    notice(e.getMessage()==null?"Processing failed":e.getMessage(),true);
                });
            }
        });
    }
    private void notice(String text,boolean error){
        if(statusText!=null){statusText.setText(text);statusText.setTextColor(error?RED:MUTED);}
        else Toast.makeText(this,text,Toast.LENGTH_SHORT).show();
    }
    private void preset(String which){
        recipe.clear();
        switch(which){
            case "json":
                sourceText.setText("{\"project\":\"RESTLESS 42\",\"status\":\"native\",\"operations\":[\"harvest\",\"index\",\"plan\"]}");
                recipe.add(new ToolboxEngine.Step("json-pretty","",true));break;
            case "jwt":
                sourceText.setText("eyJhbGciOiJub25lIn0.eyJzdWIiOiJkZW1vIn0.");
                recipe.add(new ToolboxEngine.Step("jwt-inspect","",true));break;
            case "openapi":
                sourceText.setText("{\"openapi\":\"3.0.3\",\"info\":{\"title\":\"Demo Pets\",\"version\":\"1\"},\"paths\":{\"/pets\":{\"get\":{\"operationId\":\"listPets\",\"summary\":\"List pets\"}},\"/pets/{petId}\":{\"get\":{\"operationId\":\"findPetById\",\"summary\":\"Find pet by ID\",\"parameters\":[{\"name\":\"petId\",\"in\":\"path\",\"required\":true}]}}}}");
                recipe.add(new ToolboxEngine.Step("openapi-plan","find pet by id",true));break;
        }
        renderSteps();bake();
    }
    private void paste(){
        ClipboardManager mgr=(ClipboardManager)getSystemService(CLIPBOARD_SERVICE);
        if(mgr!=null&&mgr.hasPrimaryClip()){
            ClipData data=mgr.getPrimaryClip();
            if(data!=null&&data.getItemCount()>0){
                CharSequence clip=data.getItemAt(0).coerceToText(this);
                if(clip!=null)sourceText.setText(clip.toString());
            }
        }
    }
    private void copy(){
        ClipboardManager clipboard=(ClipboardManager)getSystemService(CLIPBOARD_SERVICE);
        if(clipboard!=null)clipboard.setPrimaryClip(ClipData.newPlainText("RESTLESS 42 output",currentOutput));
        notice("Output copied. Clipboard history may be visible to other apps.",false);
    }
    private void openText(int request,String mime){
        Intent pick=new Intent(Intent.ACTION_OPEN_DOCUMENT);
        pick.setType(mime);pick.addCategory(Intent.CATEGORY_OPENABLE);
        startActivityForResult(pick,request);
    }
    private void saveText(int request,String name){
        Intent create=new Intent(Intent.ACTION_CREATE_DOCUMENT);create.addCategory(Intent.CATEGORY_OPENABLE);
        create.setType(request==SAVE_RECIPE?"application/json":"text/plain");
        create.putExtra(Intent.EXTRA_TITLE,name);startActivityForResult(create,request);
    }
    @Override protected void onActivityResult(int request,int result,Intent data){
        super.onActivityResult(request,result,data);
        if(result!=RESULT_OK||data==null||data.getData()==null)return;
        Uri uri=data.getData();
        try{
            if(request==READ_INPUT||request==IMPORT_RECIPE){
                String text=readBounded(uri);
                if(request==READ_INPUT){if(sourceText!=null)sourceText.setText(text);}
                else {recipe.clear();recipe.addAll(ToolboxEngine.importRecipe(text));renderSteps();queueBake();}
                notice("Imported local text. No upload performed.",false);
            }else if(request==SAVE_OUTPUT||request==SAVE_RECIPE){
                String payload=request==SAVE_OUTPUT?currentOutput:ToolboxEngine.exportRecipe(recipe).toString(2);
                try(OutputStream output=getContentResolver().openOutputStream(uri)){
                    if(output==null)throw new IllegalArgumentException("Could not open destination");
                    output.write(payload.getBytes(StandardCharsets.UTF_8));
                }
                notice("Saved using Android document picker.",false);
            }
        }catch(Exception e){notice("File operation failed: "+e.getMessage(),true);}
    }
    private String readBounded(Uri uri)throws Exception{
        try(InputStream stream=getContentResolver().openInputStream(uri)){
            if(stream==null)throw new IllegalArgumentException("Document unavailable");
            ByteArrayOutputStream out=new ByteArrayOutputStream();
            byte[] chunk=new byte[8192];int n;
            while((n=stream.read(chunk))>=0){out.write(chunk,0,n);if(out.size()>2_000_000)throw new IllegalArgumentException("File exceeds 2 MB");}
            return StandardCharsets.UTF_8.newDecoder()
                .onMalformedInput(java.nio.charset.CodingErrorAction.REPORT)
                .decode(java.nio.ByteBuffer.wrap(out.toByteArray())).toString();
        }
    }
    private void guideParagraph(LinearLayout main,String header,String content){
        TextView title=text(header,17,WHITE,true);main.addView(title,lp(-1,-2,8));
        TextView body=text(content,13,MUTED,false);body.setLineSpacing(dp(3),1);main.addView(body,lp(-1,-2,20));
    }
    private void showGuide(){
        LinearLayout box=paddedContent();
        heading(box,"The field guide.","Native tools. Documented behavior. Evidence before execution.");
        guideParagraph(box,"1  /  Build a recipe","Select a local operation, tap + Add, arrange steps with arrows and press RUN RECIPE. The output of one step becomes the input of the next. Up to 24 steps.");
        guideParagraph(box,"2  /  Work with documents","Paste an OpenAPI 3 JSON specification and select OpenAPI: offline probe plan. The app extracts only documented operations. There are no HTTP requests.");
        guideParagraph(box,"3  /  Keep secrets local","The app has no INTERNET permission, no WebView, no ads and no analytics. Source text remains in memory unless you explicitly use Android's document picker to export it.");
        guideParagraph(box,"4  /  Understand limits","JWT inspection is NOT signature verification. OpenAPI JSON does not prove the endpoint is deployed. A documented GET can still have unexpected effects. A plan is not authorization to probe.");
        addButton(box,action("Open full English manual ↗",true,()->openExternal("https://bspippi1337.github.io/restless/guide/")));
        addButton(box,action("Open browser toolbox ↗",false,()->openExternal("https://bspippi1337.github.io/restless/toolbox/")));
        content.addView(scroller(box));
    }
    private void showAbout(){
        LinearLayout box=paddedContent();
        heading(box,"RESTLESS 42","Native Android developer utility, debug preview 0.1.0.");
        guideParagraph(box,"Built for ownership","A practical API workbench that starts with documentation. 32 transformations implemented in native Java, with recipe support and offline OpenAPI 3 JSON operations.");
        guideParagraph(box,"Sane by default","This APK has no permissions for Internet, storage or tracking. Clipboard access and document import/export only happen when you choose those actions.");
        guideParagraph(box,"Open source","Source lives in the RESTLESS repository. Debug APKs are for sideload testing; no claim of Play Store approval or independent third-party security audit is made.");
        addButton(box,action("Project source ↗",true,()->openExternal("https://github.com/bspippi1337/restless")));
        addButton(box,action("Security policy ↗",false,()->openExternal("https://github.com/bspippi1337/restless/blob/main/SECURITY.md")));
        content.addView(scroller(box));
    }
    private void openExternal(String url){
        try{startActivity(new Intent(Intent.ACTION_VIEW,Uri.parse(url)));}
        catch(Exception e){new AlertDialog.Builder(this).setMessage("No browser was found. URL:\n"+url).setPositiveButton("OK",null).show();}
    }
}