package org.restless42.app;

import android.app.Activity;
import android.app.AlertDialog;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.content.Intent;
import android.graphics.Canvas;
import android.graphics.Color;
import android.graphics.Paint;
import android.graphics.Path;
import android.graphics.RectF;
import android.net.Uri;
import android.os.Bundle;
import android.text.InputType;
import android.view.Gravity;
import android.view.MotionEvent;
import android.view.View;
import android.view.ViewGroup;
import android.widget.Button;
import android.widget.EditText;
import android.widget.LinearLayout;
import android.widget.ScrollView;
import android.widget.TextView;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import java.net.URI;
import java.nio.ByteBuffer;
import java.nio.charset.CharacterCodingException;
import java.nio.charset.CodingErrorAction;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;

/** Fully native, entirely offline API Atlas companion to the Go documentation crawler. */
public final class AtlasActivity extends Activity {
    private static final int CHOOSE_LIBRARY=340;
    private static final int BG=Color.rgb(11,15,22),SURFACE=Color.rgb(23,32,45),
      GREEN=Color.rgb(183,247,104),LIGHT=Color.rgb(234,244,250),
      SUB=Color.rgb(157,178,195),BORDER=Color.rgb(72,92,111);
    private EditText domain;
    private TextView result,statistics;
    private LinearLayout operations;
    private ResourceMapView graph;
    private JSONArray documented=new JSONArray();
    private String host="",sourceName="";
    private final LinkedHashMap<String,List<JSONObject>> resources=new LinkedHashMap<>();

    @Override public void onCreate(Bundle state){
        super.onCreate(state);
        getWindow().setStatusBarColor(BG);
        getWindow().setNavigationBarColor(BG);
        LinearLayout content=col();
        content.setPadding(dp(15),dp(18),dp(15),dp(30));
        content.setBackgroundColor(BG);
        Button back=button("← Workbench",false,()->finish());
        content.addView(back,params(-2,44,12));
        content.addView(label("API ATLAS / NATIVE",11,GREEN,true),params(-1,-2,6));
        content.addView(label("Map the API.\nTrust the evidence.",27,LIGHT,true),params(-1,-2,10));
        content.addView(label("Domain preview, native resource map, Swagger-style operation details and request examples. Import library.json from the real RESTLESS crawler. No networking permission.",13,SUB,false),params(-1,-2,20));
        content.addView(label("01 / DOMAIN",12,GREEN,true),params(-1,-2,7));
        domain=new EditText(this);domain.setSingleLine(true);
        domain.setTextColor(LIGHT);domain.setHintTextColor(SUB);domain.setTextSize(13);
        domain.setHint("docs.example.org or exact HTTPS docs URL");
        domain.setInputType(InputType.TYPE_CLASS_TEXT|InputType.TYPE_TEXT_VARIATION_URI);
        domain.setBackgroundTintList(android.content.res.ColorStateList.valueOf(GREEN));
        content.addView(domain,params(-1,52,10));
        LinearLayout actions=row();
        actions.addView(button("Preview scope",true,this::preview),params(-2,43,0));
        actions.addView(button("Import library.json",false,this::open),params(-2,43,0));
        content.addView(actions,params(-1,-2,8));
        result=label("Ready. No network access. Domain previews are local.",12,SUB,false);
        result.setPadding(dp(13),dp(12),dp(13),dp(12));
        result.setBackgroundColor(SURFACE);
        content.addView(result,params(-1,-2,23));
        content.addView(label("02 / SOURCE-BACKED MAP",12,GREEN,true),params(-1,-2,8));
        statistics=label("Import a RESTLESS library.json to map its documented operations.",12,SUB,false);
        content.addView(statistics,params(-1,-2,8));
        graph=new ResourceMapView();
        content.addView(graph,params(-1,190,13));
        content.addView(label("03 / DOCUMENTED API OPERATIONS",12,GREEN,true),params(-1,-2,9));
        operations=col();
        content.addView(operations,params(-1,-2,12));
        showEmpty();
        ScrollView scroll=new ScrollView(this);
        scroll.setFillViewport(true);
        scroll.addView(content,new ScrollView.LayoutParams(-1,-2));
        setContentView(scroll);
    }
    private int dp(float x){return (int)(getResources().getDisplayMetrics().density*x+.5f);}
    private LinearLayout col(){LinearLayout a=new LinearLayout(this);a.setOrientation(LinearLayout.VERTICAL);return a;}
    private LinearLayout row(){LinearLayout a=new LinearLayout(this);a.setOrientation(LinearLayout.HORIZONTAL);a.setGravity(Gravity.CENTER_VERTICAL);return a;}
    private LinearLayout.LayoutParams params(int w,int h,int bottom){
        LinearLayout.LayoutParams p=new LinearLayout.LayoutParams(w<0?w:dp(w),h<0?h:dp(h));
        p.bottomMargin=dp(bottom);p.rightMargin=dp(7);return p;
    }
    private TextView label(String text,int size,int color,boolean bold){
        TextView t=new TextView(this);t.setText(text);t.setTextSize(size);t.setTextColor(color);
        if(bold)t.setTypeface(null,1);
        t.setLineSpacing(dp(3),1);return t;
    }
    private Button button(String text,boolean primary,Runnable action){
        Button b=new Button(this);b.setText(text);b.setAllCaps(false);b.setTextSize(12);
        b.setTextColor(primary?BG:GREEN);
        b.setOnClickListener(v->action.run());
        return b;
    }
    private void preview(){
        try{
            String raw=domain.getText().toString().trim();
            if(raw.isEmpty()||raw.length()>2048||raw.matches(".*\\s+.*"))throw new IllegalArgumentException("Enter one public DNS hostname.");
            if(!raw.contains("://"))raw="https://"+raw;
            URI uri=new URI(raw);
            if(!"https".equalsIgnoreCase(uri.getScheme())||uri.getHost()==null||
               uri.getRawUserInfo()!=null||uri.getRawQuery()!=null||uri.getRawFragment()!=null||uri.getPort()!=-1)
                throw new IllegalArgumentException("HTTPS only, without credentials, query, fragment or custom port.");
            String name=uri.getHost().toLowerCase(Locale.ROOT);
            if(name.matches("\\d+(?:\\.\\d+){3}")||!name.matches("[a-z0-9][a-z0-9.-]*[a-z0-9]")||
                !name.contains(".")||name.contains("..")||
                name.endsWith(".local")||name.endsWith(".internal")||name.endsWith(".test")||
                name.endsWith(".invalid")||name.endsWith(".localhost")||name.endsWith(".onion"))
                throw new IllegalArgumentException("Local, reserved and IP hosts are not permitted.");
            String url=uri.toASCIIString();
            if("/".equals(uri.getPath())||uri.getPath().isEmpty())url="https://"+name+"/";
            String command="restless library '"+url.replace("'","")+"'"+" --max-pages 12 --max-depth 2 --execute --authorized";
            result.setText("DRY RUN / ZERO REQUESTS\nExact host: "+name+"\nSeed: "+url+
                "\n\nRun the real, authorized crawler in RESTLESS desktop/CLI, then import its library.json here.");
            final String copy=command;
            new AlertDialog.Builder(this)
                .setTitle("Documentation crawl command")
                .setMessage(copy+"\n\nExecuted only by the native CLI after you check site permissions. This Android screen sends no network request.")
                .setPositiveButton("Copy command",(d,w)->{
                    ClipboardManager cm=(ClipboardManager)getSystemService(CLIPBOARD_SERVICE);
                    if(cm!=null)cm.setPrimaryClip(ClipData.newPlainText("RESTLESS 42 CLI",copy));
                })
                .setNegativeButton("Close",null).show();
        }catch(Exception e){result.setText("Domain rejected: "+e.getMessage());}
    }
    private void open(){
        Intent intent=new Intent(Intent.ACTION_OPEN_DOCUMENT);
        intent.addCategory(Intent.CATEGORY_OPENABLE);
        intent.setType("application/json");
        startActivityForResult(intent,CHOOSE_LIBRARY);
    }
    @Override protected void onActivityResult(int req,int resultCode,Intent data){
        super.onActivityResult(req,resultCode,data);
        if(req!=CHOOSE_LIBRARY||resultCode!=RESULT_OK||data==null||data.getData()==null)return;
        try{
            String text=readBounded(data.getData());
            JSONObject root=new JSONObject(text);
            if(root.optInt("schema_version",-1)!=1||root.optJSONObject("target")==null||root.optJSONArray("operations")==null)
                throw new IllegalArgumentException("Expected native RESTLESS library.json schema v1.");
            String h=root.getJSONObject("target").optString("host");
            if(h.isEmpty())throw new IllegalArgumentException("Library hostname is missing.");
            JSONArray list=root.getJSONArray("operations");
            if(list.length()>1000)throw new IllegalArgumentException("Maximum 1,000 documented operations.");
            host=h;sourceName=data.getData().getLastPathSegment();
            documented=list;resources.clear();
            for(int i=0;i<list.length();i++){
                JSONObject op=list.getJSONObject(i);
                String path=op.optString("path"),method=op.optString("method");
                if(!path.startsWith("/")||!method.matches("GET|POST|PATCH|PUT|DELETE|HEAD|OPTIONS|TRACE"))continue;
                if(!resources.containsKey(path))resources.put(path,new ArrayList<>());
                resources.get(path).add(op);
            }
            int docs=root.optJSONArray("documents")==null?0:root.getJSONArray("documents").length();
            statistics.setText(host+"  /  "+docs+" evidence documents  /  "+list.length()+" declared operations  /  "+resources.size()+" paths");
            graph.setItems(host,new ArrayList<>(resources.keySet()));
            showOperations();
            result.setText("IMPORTED LOCALLY\n"+docs+" documents, "+list.length()+" documented HTTP methods.\nNo endpoints guessed or executed.");
        }catch(Exception e){result.setText("Import rejected: "+e.getMessage());}
    }
    private String readBounded(Uri file)throws Exception{
        try(InputStream in=getContentResolver().openInputStream(file)){
            if(in==null)throw new IllegalArgumentException("Unable to open file");
            ByteArrayOutputStream out=new ByteArrayOutputStream();
            byte[] chunk=new byte[8192];int count;
            while((count=in.read(chunk))!=-1){
                out.write(chunk,0,count);
                if(out.size()>2_000_000)throw new IllegalArgumentException("Library exceeds 2 MB");
            }
            return StandardCharsets.UTF_8.newDecoder()
                .onMalformedInput(CodingErrorAction.REPORT)
                .decode(ByteBuffer.wrap(out.toByteArray())).toString();
        }
    }
    private void showEmpty(){
        operations.removeAllViews();
        TextView t=label("The map is empty until a real library is imported.",13,SUB,false);
        t.setPadding(dp(12),dp(18),dp(12),dp(18));
        operations.addView(t,params(-1,-2,0));
    }
    private void showOperations(){
        operations.removeAllViews();
        for(Map.Entry<String,List<JSONObject>> resource:resources.entrySet()){
            TextView head=label(resource.getKey(),14,GREEN,true);
            head.setPadding(dp(4),dp(9),dp(4),dp(6));
            operations.addView(head,params(-1,-2,2));
            for(JSONObject op:resource.getValue()){
                String method=op.optString("method"),title=op.optString("summary",op.optString("operation_id"));
                TextView entry=label(method+"     "+title,13,LIGHT,false);
                entry.setBackgroundColor(SURFACE);
                entry.setPadding(dp(13),dp(12),dp(13),dp(12));
                entry.setOnClickListener(v->showOperation(op));
                entry.setContentDescription("Inspect documented "+method+" "+resource.getKey());
                operations.addView(entry,params(-1,-2,6));
            }
        }
        if(resources.isEmpty())showEmpty();
    }
    private void showOperation(JSONObject op){
        String method=op.optString("method"),path=op.optString("path");
        StringBuilder text=new StringBuilder("DOCUMENTATION ONLY\n\n"+method+" "+path+"\n");
        text.append(op.optString("summary")).append("\n\nINPUT PARAMETERS\n");
        JSONArray params=op.optJSONArray("parameters");
        if(params==null||params.length()==0)text.append("None declared\n");
        else for(int i=0;i<params.length();i++){
            JSONObject p=params.optJSONObject(i);if(p==null)continue;
            JSONObject schema=p.optJSONObject("schema");
            String type=schema==null?"string":schema.optString("type","string");
            text.append("- ").append(p.optString("name")).append(" : ").append(type)
                .append(" (").append(p.optString("in")).append(p.optBoolean("required")?", required":"").append(")\n");
        }
        text.append("\nAUTH DECLARATIONS\n").append(op.optJSONArray("security")).append("\n");
        JSONObject request=op.optJSONObject("request");
        if(request!=null)text.append("\nREQUEST BODY EXAMPLE (NOT EXECUTED)\n").append(request.toString()).append("\n");
        JSONArray responses=op.optJSONArray("responses");
        if(responses!=null)text.append("\nDOCUMENTED RESPONSES\n").append(responses.toString()).append("\n");
        text.append("\nREQUEST PREVIEW (NEVER EXECUTED)\n");
        text.append("curl --request ").append(method).append(" --url 'https://")
            .append(host).append(path).append("' --header 'Accept: application/json'\n");
        text.append("\nSOURCE\n").append(op.optString("source_url"))
            .append("\nSHA256: ").append(op.optString("source_sha256","not available"));
        text.append("\n\nNo live API availability or authorization has been verified.");
        final String preview=text.toString();
        new AlertDialog.Builder(this)
            .setTitle(method+" "+path)
            .setMessage(preview)
            .setPositiveButton("Copy preview",(d,w)->{
                ClipboardManager cm=(ClipboardManager)getSystemService(CLIPBOARD_SERVICE);
                if(cm!=null)cm.setPrimaryClip(ClipData.newPlainText("RESTLESS API operation",preview));
            })
            .setNegativeButton("Close",null).show();
    }
    /** Canvas-drawn resource graph, no HTML, WebView or remote assets. */
    private final class ResourceMapView extends View {
        private final Paint paint=new Paint(3);
        private List<String> paths=new ArrayList<>();
        private String domainName="No library";
        ResourceMapView(){super(AtlasActivity.this);setLayerType(View.LAYER_TYPE_SOFTWARE,null);}
        void setItems(String domain,List<String> p){domainName=domain;paths=p;invalidate();requestLayout();}
        @Override protected void onMeasure(int w,int h){
            int width=MeasureSpec.getSize(w),height=dp(Math.max(190,95+paths.size()*52));
            setMeasuredDimension(width,height);
        }
        @Override protected void onDraw(Canvas c){
            super.onDraw(c);
            float width=getWidth(),margin=dp(8),rootX=dp(10),rootY=dp(10);
            paint.setStyle(Paint.Style.FILL);
            paint.setColor(SURFACE);
            c.drawRoundRect(new RectF(rootX,rootY,width-margin,dp(63)),dp(9),dp(9),paint);
            paint.setColor(GREEN);paint.setTextSize(dp(15));paint.setFakeBoldText(true);
            c.drawText(domainName.length()>34?domainName.substring(0,31)+"...":domainName,dp(23),dp(32),paint);
            paint.setColor(SUB);paint.setTextSize(dp(10));paint.setFakeBoldText(false);
            c.drawText("DOCUMENTED HOST  /  "+paths.size()+" resources",dp(23),dp(49),paint);
            for(int i=0;i<paths.size();i++){
                float top=dp(85+i*52),left=dp(28);
                paint.setColor(BORDER);paint.setStyle(Paint.Style.STROKE);paint.setStrokeWidth(dp(1.3f));
                c.drawLine(dp(27),dp(63),dp(27),top+dp(18),paint);
                c.drawLine(dp(27),top+dp(18),left+dp(13),top+dp(18),paint);
                paint.setStyle(Paint.Style.FILL);paint.setColor(SURFACE);
                c.drawRoundRect(new RectF(left,top,width-margin,top+dp(41)),dp(8),dp(8),paint);
                paint.setColor(LIGHT);paint.setTextSize(dp(12));paint.setFakeBoldText(true);
                String name=paths.get(i);
                c.drawText(name.length()>42?name.substring(0,39)+"...":name,left+dp(11),top+dp(24),paint);
            }
        }
    }
}
