package org.restless42.app;

import android.util.Base64;
import org.json.JSONArray;
import org.json.JSONObject;
import org.json.JSONTokener;

import java.net.URI;
import java.net.URLDecoder;
import java.net.URLEncoder;
import java.nio.ByteBuffer;
import java.nio.charset.CharacterCodingException;
import java.nio.charset.CodingErrorAction;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.Comparator;
import java.util.LinkedHashMap;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.Set;

/**
 * RESTLESS 42 offline transformation engine.
 * No HTTP client, WebView, telemetry, file access, or Android service dependencies.
 */
public final class ToolboxEngine {
    public static final class Tool {
        public final String id;
        public final String label;
        public final String category;
        public final String optionLabel;
        public final String defaultOption;
        Tool(String id, String label, String category) { this(id,label,category,"",""); }
        Tool(String id, String label, String category, String optionLabel, String defaultOption) {
            this.id=id; this.label=label; this.category=category; this.optionLabel=optionLabel; this.defaultOption=defaultOption;
        }
        @Override public String toString() { return label; }
    }
    public static final class Step {
        public String id, option;
        public boolean enabled;
        Step(String id, String option, boolean enabled) { this.id=id; this.option=option; this.enabled=enabled; }
        JSONObject json() throws Exception { return new JSONObject().put("id",id).put("option",option).put("enabled",enabled); }
    }
    public static final List<Tool> TOOLS = Collections.unmodifiableList(Arrays.asList(
        new Tool("json-pretty","JSON: pretty print","JSON"),
        new Tool("json-minify","JSON: minify","JSON"),
        new Tool("json-sort","JSON: sort keys","JSON"),
        new Tool("json-path","JSON: extract path","JSON","Path","items[0].name"),
        new Tool("json-to-csv","JSON array to CSV","JSON"),
        new Tool("csv-to-json","CSV to JSON","JSON"),
        new Tool("base64-encode","To Base64","Encoding"),
        new Tool("base64-decode","From Base64","Encoding"),
        new Tool("base64url-encode","To Base64URL","Encoding"),
        new Tool("base64url-decode","From Base64URL","Encoding"),
        new Tool("url-encode","URL component encode","Encoding"),
        new Tool("url-decode","URL component decode","Encoding"),
        new Tool("hex-encode","UTF-8 to Hex","Encoding"),
        new Tool("hex-decode","Hex to UTF-8","Encoding"),
        new Tool("html-escape","Escape HTML","Encoding"),
        new Tool("html-unescape","Unescape HTML","Encoding"),
        new Tool("jwt-inspect","JWT: inspect unverified","Security"),
        new Tool("sha-256","SHA-256 digest","Security"),
        new Tool("sha-512","SHA-512 digest","Security"),
        new Tool("rot13","ROT13","Text"),
        new Tool("uppercase","Uppercase","Text"),
        new Tool("lowercase","Lowercase","Text"),
        new Tool("trim","Trim whitespace","Text"),
        new Tool("sort-lines","Sort lines","Text"),
        new Tool("unique-lines","Unique lines","Text"),
        new Tool("text-stats","Text statistics","Text"),
        new Tool("url-inspect","Inspect URL","Web"),
        new Tool("query-parse","Query string to JSON","Web"),
        new Tool("query-stringify","JSON to query string","Web"),
        new Tool("openapi-operations","OpenAPI: list operations","OpenAPI"),
        new Tool("openapi-search","OpenAPI: find operation","OpenAPI","Keywords","find pet by id"),
        new Tool("openapi-plan","OpenAPI: offline probe plan","OpenAPI","Task","find pet by id")
    ));
    private static final int MAX_CHARS = 2_000_000, MAX_STEPS = 24, MAX_OPERATIONS = 1000;
    private static final Set<String> METHODS = new LinkedHashSet<>(Arrays.asList(
        "get","head","options","post","put","patch","delete","trace"
    ));
    private ToolboxEngine() {}
    public static Tool find(String id) {
        for(Tool t:TOOLS) if(t.id.equals(id)) return t;
        return null;
    }
    public static String run(String source, List<Step> recipe) throws Exception {
        check(source);
        if(recipe.size()>MAX_STEPS)throw new IllegalArgumentException("Recipe exceeds 24 steps");
        String text=source;
        for(int i=0;i<recipe.size();i++){
            Step step=recipe.get(i);
            Tool tool=find(step.id);
            if(tool==null)throw new IllegalArgumentException("Unknown operation at step "+(i+1));
            if(!step.enabled)continue;
            try { text=apply(step.id,text,step.option==null?"":step.option); check(text); }
            catch(Exception e) { throw new IllegalArgumentException("Step "+(i+1)+" ("+tool.label+"): "+e.getMessage(),e); }
        }
        return text;
    }
    public static JSONObject exportRecipe(List<Step> recipe) throws Exception {
        if(recipe.size()>MAX_STEPS) throw new IllegalArgumentException("Too many steps");
        JSONArray items=new JSONArray();
        for(Step step:recipe){if(find(step.id)==null)throw new IllegalArgumentException("Unknown operation");items.put(step.json());}
        return new JSONObject().put("format","restless-42-toolbox-recipe").put("version",1).put("steps",items);
    }
    public static List<Step> importRecipe(String input) throws Exception {
        JSONObject obj=new JSONObject(input);
        if(!"restless-42-toolbox-recipe".equals(obj.optString("format"))||obj.optInt("version")!=1)throw new IllegalArgumentException("Unsupported recipe");
        JSONArray items=obj.getJSONArray("steps");
        if(items.length()>MAX_STEPS)throw new IllegalArgumentException("Recipe exceeds 24 steps");
        List<Step> steps=new ArrayList<>();
        for(int i=0;i<items.length();i++){
            JSONObject item=items.getJSONObject(i);
            String id=item.getString("id"),option=item.optString("option","");
            if(find(id)==null||option.length()>500)throw new IllegalArgumentException("Invalid operation at step "+(i+1));
            steps.add(new Step(id,option,item.optBoolean("enabled",true)));
        }
        return steps;
    }
    private static void check(String value) {
        if(value==null||value.length()>MAX_CHARS)throw new IllegalArgumentException("Maximum text size: 2 million characters");
    }
    private static String utf8(byte[] bytes) throws Exception {
        try{return StandardCharsets.UTF_8.newDecoder().onMalformedInput(CodingErrorAction.REPORT)
            .onUnmappableCharacter(CodingErrorAction.REPORT).decode(ByteBuffer.wrap(bytes)).toString();}
        catch(CharacterCodingException e){throw new IllegalArgumentException("Bytes are not valid UTF-8",e);}
    }
    private static String primitive(Object value) {
        if(value==null||value==JSONObject.NULL)return "null";
        if(value instanceof String)return JSONObject.quote((String)value);
        if(value instanceof Number||value instanceof Boolean)return String.valueOf(value);
        throw new IllegalArgumentException("Unsupported primitive JSON value");
    }
    private static String pretty(Object value) throws Exception {
        if(value instanceof JSONObject)return ((JSONObject)value).toString(2);
        if(value instanceof JSONArray)return ((JSONArray)value).toString(2);
        return primitive(value);
    }
    private static Object json(String value) throws Exception {
        JSONTokener parser=new JSONTokener(value);
        Object o=parser.nextValue();
        if(parser.nextClean()!=0)throw new IllegalArgumentException("Trailing data after JSON value");
        return o;
    }
    private static String canonical(Object value) throws Exception {
        if(value instanceof JSONObject){
            JSONObject o=(JSONObject)value; List<String> keys=new ArrayList<>();
            java.util.Iterator<String> it=o.keys();while(it.hasNext())keys.add(it.next());
            Collections.sort(keys);List<String> segments=new ArrayList<>();
            for(String key:keys)segments.add(JSONObject.quote(key)+":"+canonical(o.get(key)));
            return "{"+String.join(",",segments)+"}";
        }
        if(value instanceof JSONArray){
            JSONArray a=(JSONArray)value;List<String> segments=new ArrayList<>();
            for(int i=0;i<a.length();i++)segments.add(canonical(a.get(i)));
            return "["+String.join(",",segments)+"]";
        }
        return primitive(value);
    }
    private static String hex(byte[] bytes) {
        final char[] digits="0123456789abcdef".toCharArray();char[] out=new char[bytes.length*2];
        for(int i=0;i<bytes.length;i++){int b=bytes[i]&255;out[i*2]=digits[b>>>4];out[i*2+1]=digits[b&15];}
        return new String(out);
    }
    private static byte[] hexDecode(String text){
        String s=text.replaceAll("\\s+","");
        if((s.length()&1)==1||!s.matches("[0-9a-fA-F]*"))throw new IllegalArgumentException("Hexadecimal bytes must have two digits each");
        byte[] out=new byte[s.length()/2];
        for(int i=0;i<s.length();i+=2)out[i/2]=(byte)Integer.parseInt(s.substring(i,i+2),16);
        return out;
    }
    private static String escapeHtml(String text){
        return text.replace("&","&amp;").replace("<","&lt;").replace(">","&gt;").replace("\"","&quot;").replace("'","&#39;");
    }
    private static String unescapeHtml(String text){
        return text.replace("&lt;","<").replace("&gt;",">").replace("&quot;","\"").replace("&#39;","'").replace("&apos;","'").replace("&amp;","&");
    }
    private static String rot13(String text) {
        StringBuilder b=new StringBuilder(text.length());
        for(char c:text.toCharArray()){
            if(c>='a'&&c<='z')b.append((char)('a'+(c-'a'+13)%26));
            else if(c>='A'&&c<='Z')b.append((char)('A'+(c-'A'+13)%26));
            else b.append(c);
        }
        return b.toString();
    }
    private static String extractPath(String text,String path)throws Exception{
        if(path.trim().isEmpty())throw new IllegalArgumentException("Specify a JSON property path");
        Object cursor=json(text);String p=path.replaceFirst("^\\$\\.?", "");
        String[] parts=p.replace("[",".").replace("]","").split("\\.");
        for(String part:parts){
            if(part.isEmpty())continue;
            if(cursor instanceof JSONObject) cursor=((JSONObject)cursor).get(part);
            else if(cursor instanceof JSONArray)cursor=((JSONArray)cursor).get(Integer.parseInt(part));
            else throw new IllegalArgumentException("JSON property path does not exist: "+path);
        }
        return cursor instanceof String?(String)cursor:pretty(cursor);
    }
    private static String cell(Object value){
        String s=value==null||value==JSONObject.NULL?"":value.toString();
        return s.matches("(?s).*[\\\",\\r\\n].*")?"\""+s.replace("\"","\"\"")+"\"":s;
    }
    private static String toCSV(String text)throws Exception{
        Object value=json(text);if(!(value instanceof JSONArray))throw new IllegalArgumentException("Expected a JSON array of objects");
        JSONArray array=(JSONArray)value;LinkedHashSet<String> fields=new LinkedHashSet<>();
        for(int i=0;i<array.length();i++){Object row=array.get(i);if(!(row instanceof JSONObject))throw new IllegalArgumentException("CSV rows must be JSON objects");java.util.Iterator<String> it=((JSONObject)row).keys();while(it.hasNext())fields.add(it.next());}
        if(fields.isEmpty())return "";
        List<String> keys=new ArrayList<>(fields);StringBuilder output=new StringBuilder();
        for(int i=0;i<keys.size();i++){if(i>0)output.append(',');output.append(cell(keys.get(i)));}output.append("\r\n");
        for(int i=0;i<array.length();i++){JSONObject row=array.getJSONObject(i);for(int j=0;j<keys.size();j++){if(j>0)output.append(',');output.append(cell(row.opt(keys.get(j))));}if(i+1<array.length())output.append("\r\n");}
        return output.toString();
    }
    private static List<List<String>> parseCSV(String text) {
        List<List<String>> rows=new ArrayList<>();List<String> row=new ArrayList<>();StringBuilder cell=new StringBuilder();
        boolean quoted=false,justClosed=false;
        for(int i=0;i<text.length();i++){
            char ch=text.charAt(i);
            if(quoted){
                if(ch=='"'&&i+1<text.length()&&text.charAt(i+1)=='"'){cell.append('"');i++;}
                else if(ch=='"'){quoted=false;justClosed=true;}
                else cell.append(ch);
                continue;
            }
            if(justClosed){if(ch!=','&&ch!='\r'&&ch!='\n')throw new IllegalArgumentException("Invalid CSV after closing quote");justClosed=false;}
            if(ch=='"'&&cell.length()==0){quoted=true;continue;}
            if(ch=='"')throw new IllegalArgumentException("Unexpected CSV quote");
            if(ch==','){row.add(cell.toString());cell.setLength(0);continue;}
            if(ch=='\r'||ch=='\n'){row.add(cell.toString());rows.add(row);row=new ArrayList<>();cell.setLength(0);if(ch=='\r'&&i+1<text.length()&&text.charAt(i+1)=='\n')i++;if(rows.size()>5001)throw new IllegalArgumentException("CSV exceeds 5,000 rows");continue;}
            cell.append(ch);
        }
        if(quoted)throw new IllegalArgumentException("Unterminated quoted CSV field");
        if(cell.length()>0||!row.isEmpty()){row.add(cell.toString());rows.add(row);}
        while(!rows.isEmpty()&&rows.get(rows.size()-1).size()==1&&rows.get(rows.size()-1).get(0).isEmpty())rows.remove(rows.size()-1);
        return rows;
    }
    private static String fromCSV(String text)throws Exception{
        List<List<String>> rows=parseCSV(text);if(rows.isEmpty())return "[]";
        List<String> columns=rows.remove(0);if(new LinkedHashSet<>(columns).size()!=columns.size())throw new IllegalArgumentException("Duplicate CSV headers");
        JSONArray output=new JSONArray();
        for(int i=0;i<rows.size();i++){
            List<String> cells=rows.get(i);
            if(cells.size()!=columns.size())throw new IllegalArgumentException("Wrong number of CSV fields at row "+(i+2));
            JSONObject row=new JSONObject();for(int j=0;j<columns.size();j++)row.put(columns.get(j),cells.get(j));output.put(row);
        }
        return pretty(output);
    }
    private static JSONObject queryObject(String raw)throws Exception{
        JSONObject out=new JSONObject();if(raw.startsWith("?"))raw=raw.substring(1);
        if(raw.isEmpty())return out;
        for(String pair:raw.split("&",-1)){
            if(pair.isEmpty())continue;
            String[] parts=pair.split("=",2);
            String k=URLDecoder.decode(parts[0],"UTF-8"),v=URLDecoder.decode(parts.length>1?parts[1]:"","UTF-8");
            JSONArray values=out.optJSONArray(k);if(values==null){values=new JSONArray();out.put(k,values);}values.put(v);
        }
        return out;
    }
    private static String queryString(String text)throws Exception{
        Object root=json(text);if(!(root instanceof JSONObject))throw new IllegalArgumentException("Expected JSON object");
        JSONObject obj=(JSONObject)root;List<String> pairs=new ArrayList<>();java.util.Iterator<String> keys=obj.keys();
        while(keys.hasNext()){
            String key=keys.next();Object item=obj.get(key);
            if(item instanceof JSONArray){JSONArray a=(JSONArray)item;for(int i=0;i<a.length();i++)pairs.add(URLEncoder.encode(key,"UTF-8")+"="+URLEncoder.encode(a.optString(i,""),"UTF-8"));}
            else pairs.add(URLEncoder.encode(key,"UTF-8")+"="+URLEncoder.encode(item==JSONObject.NULL?"":String.valueOf(item),"UTF-8"));
        }
        return String.join("&",pairs);
    }
    private static String urlInspect(String text)throws Exception {
        URI u=new URI(text.trim());if(u.getScheme()==null||u.getHost()==null)throw new IllegalArgumentException("Expected an absolute URL with scheme");
        return pretty(new JSONObject().put("scheme",u.getScheme()).put("host",u.getHost()).put("port",u.getPort()).put("path",u.getPath())
          .put("query",queryObject(u.getRawQuery()==null?"":u.getRawQuery())).put("fragment",u.getFragment()==null?"":u.getFragment())
          .put("networkRequests",0).put("note","Parsed locally; no API call was made"));
    }
    private static JSONObject jwtInspect(String text)throws Exception{
        String[] sections=text.trim().split("\\.",-1);if(sections.length!=3)throw new IllegalArgumentException("JWT must have three dot-separated parts");
        JSONObject header=new JSONObject(utf8(Base64.decode(sections[0],Base64.URL_SAFE|Base64.NO_WRAP)));
        JSONObject payload=new JSONObject(utf8(Base64.decode(sections[1],Base64.URL_SAFE|Base64.NO_WRAP)));
        return new JSONObject().put("header",header).put("payload",payload).put("signaturePresent",!sections[2].isEmpty())
          .put("signatureVerified",false).put("warning","Inspection only; signature, audience, issuer, expiry and trust were NOT checked");
    }
    private static List<String> keywords(String text){
        LinkedHashSet<String> words=new LinkedHashSet<>();
        for(String part:text.toLowerCase(Locale.ROOT).split("[^a-z0-9]+"))if(part.length()>1)words.add(part);
        return new ArrayList<>(words);
    }
    private static JSONArray resolveParameters(JSONObject spec,JSONArray source) throws Exception{
        JSONArray result=new JSONArray();if(source==null)return result;
        for(int i=0;i<source.length();i++){
            JSONObject input=source.optJSONObject(i);if(input==null)continue;
            JSONObject p=input;Set<String> visited=new LinkedHashSet<>();
            for(int depth=0;depth<8&&p.has("$ref");depth++){
                String ref=p.optString("$ref");String prefix="#/components/parameters/";
                if(!ref.startsWith(prefix)||!visited.add(ref))break;
                String name=ref.substring(prefix.length()).replace("~1","/").replace("~0","~");
                JSONObject next=spec.optJSONObject("components");
                next=next==null?null:next.optJSONObject("parameters");
                p=next==null?null:next.optJSONObject(name);
                if(p==null){p=input;break;}
            }
            JSONObject entry=new JSONObject();
            entry.put("name",p.optString("name"));entry.put("in",p.optString("in"));entry.put("required",p.optBoolean("required",false));
            entry.put("unresolvedRef",p.has("$ref"));result.put(entry);
        }
        return result;
    }
    private static JSONArray operations(String text)throws Exception {
        JSONObject spec=new JSONObject(text);
        if(!spec.optString("openapi").startsWith("3."))throw new IllegalArgumentException("Expected an OpenAPI 3.x JSON document");
        JSONObject paths=spec.optJSONObject("paths");if(paths==null)throw new IllegalArgumentException("Specification has no paths object");
        JSONArray ops=new JSONArray();java.util.Iterator<String> it=paths.keys();
        while(it.hasNext()){
            String path=it.next();if(!path.startsWith("/"))continue;
            JSONObject pathItem=paths.optJSONObject(path);if(pathItem==null||pathItem.has("$ref"))continue;
            for(String method:METHODS){
                JSONObject op=pathItem.optJSONObject(method);if(op==null)continue;
                JSONArray parents=resolveParameters(spec,pathItem.optJSONArray("parameters"));
                JSONArray children=resolveParameters(spec,op.optJSONArray("parameters"));
                LinkedHashMap<String,JSONObject> merged=new LinkedHashMap<>();
                for(JSONArray a:Arrays.asList(parents,children))for(int i=0;i<a.length();i++){
                    JSONObject param=a.getJSONObject(i);
                    merged.put(param.optString("in")+":"+param.optString("name"),param);
                }
                JSONArray params=new JSONArray();boolean unresolved=false,required=false;
                for(JSONObject p:merged.values()){params.put(p);unresolved|=p.optBoolean("unresolvedRef");required|=p.optBoolean("required");}
                unresolved|=op.optJSONObject("requestBody")!=null && op.optJSONObject("requestBody").has("$ref");
                JSONArray security=op.has("security")?op.optJSONArray("security"):spec.optJSONArray("security");
                if(security==null)security=new JSONArray();
                boolean review=!method.equals("get")&&!method.equals("head")||security.length()>0||path.contains("{")||required||unresolved;
                ops.put(new JSONObject().put("method",method.toUpperCase(Locale.ROOT)).put("path",path)
                    .put("operationId",op.optString("operationId")).put("summary",op.optString("summary"))
                    .put("description",op.optString("description")).put("tags",op.optJSONArray("tags")==null?new JSONArray():op.optJSONArray("tags"))
                    .put("parameters",params).put("security",security).put("unresolvedRefs",unresolved).put("reviewRequired",review));
                if(ops.length()>MAX_OPERATIONS)throw new IllegalArgumentException("Maximum 1000 documented operations");
            }
        }
        return ops;
    }
    private static JSONArray rank(JSONArray operations,String task)throws Exception{
        List<String> wanted=keywords(task);if(wanted.isEmpty())throw new IllegalArgumentException("Enter search keywords");
        List<JSONObject> scored=new ArrayList<>();
        for(int i=0;i<operations.length();i++){
            JSONObject op=operations.getJSONObject(i);int score=0;
            for(String[] field:new String[][]{{"operationId","8"},{"path","6"},{"summary","6"},{"description","2"},{"method","3"}}){
                List<String> tokens=keywords(op.optString(field[0]));for(String term:wanted)for(String token:tokens)if(token.equals(term)||term.length()>3&&token.startsWith(term)){score+=Integer.parseInt(field[1]);break;}
            }
            JSONArray params=op.optJSONArray("parameters");
            if(params!=null)for(int j=0;j<params.length();j++){
                List<String> tokens=keywords(params.getJSONObject(j).optString("name"));
                for(String term:wanted)if(tokens.contains(term))score+=2;
            }
            if(score==0)continue;
            if("GET".equals(op.optString("method"))||"HEAD".equals(op.optString("method")))score+=4;
            op.put("relevanceScore",score);scored.add(op);
        }
        scored.sort(Comparator.comparingInt((JSONObject x)->x.optInt("relevanceScore")).reversed()
            .thenComparing(x->x.optString("path")).thenComparing(x->x.optString("method")));
        JSONArray output=new JSONArray();for(int i=0;i<Math.min(10,scored.size());i++)output.put(scored.get(i));
        return output;
    }
    private static JSONObject plan(String spec,String task)throws Exception{
        JSONArray hits=rank(operations(spec),task);
        return new JSONObject().put("mode","DOCUMENTATION_ONLY_NO_EXECUTION").put("networkRequests",0)
            .put("task",task).put("source","User-supplied JSON OpenAPI 3 specification; deployment not verified")
            .put("candidates",hits)
            .put("warning","A documented API operation is not permission to call it or proof it is deployed. Review credentials and possible effects.");
    }
    public static String apply(String id,String source,String option)throws Exception{
        check(source);if(option==null)option="";
        switch(id){
            case "json-pretty":return pretty(json(source));
            case "json-minify":return json(source) instanceof JSONObject?((JSONObject)json(source)).toString():json(source) instanceof JSONArray?((JSONArray)json(source)).toString():pretty(json(source));
            case "json-sort":return canonical(json(source));
            case "json-path":return extractPath(source,option);
            case "json-to-csv":return toCSV(source);
            case "csv-to-json":return fromCSV(source);
            case "base64-encode":return Base64.encodeToString(source.getBytes(StandardCharsets.UTF_8),Base64.NO_WRAP);
            case "base64-decode":return utf8(Base64.decode(source.trim(),Base64.DEFAULT));
            case "base64url-encode":return Base64.encodeToString(source.getBytes(StandardCharsets.UTF_8),Base64.URL_SAFE|Base64.NO_PADDING|Base64.NO_WRAP);
            case "base64url-decode":return utf8(Base64.decode(source.trim(),Base64.URL_SAFE|Base64.NO_WRAP));
            case "url-encode":return URLEncoder.encode(source,"UTF-8").replace("+","%20");
            case "url-decode":return URLDecoder.decode(source,"UTF-8");
            case "hex-encode":return hex(source.getBytes(StandardCharsets.UTF_8));
            case "hex-decode":return utf8(hexDecode(source));
            case "html-escape":return escapeHtml(source);
            case "html-unescape":return unescapeHtml(source);
            case "jwt-inspect":return pretty(jwtInspect(source));
            case "sha-256":return hex(MessageDigest.getInstance("SHA-256").digest(source.getBytes(StandardCharsets.UTF_8)));
            case "sha-512":return hex(MessageDigest.getInstance("SHA-512").digest(source.getBytes(StandardCharsets.UTF_8)));
            case "rot13":return rot13(source);
            case "uppercase":return source.toUpperCase(Locale.ROOT);
            case "lowercase":return source.toLowerCase(Locale.ROOT);
            case "trim":return source.trim();
            case "sort-lines":{List<String> lines=new ArrayList<>(Arrays.asList(source.split("\\r?\\n",-1)));Collections.sort(lines);return String.join("\n",lines);}
            case "unique-lines":return String.join("\n",new LinkedHashSet<>(Arrays.asList(source.split("\\r?\\n",-1))));
            case "text-stats":return pretty(new JSONObject().put("characters",source.codePointCount(0,source.length()))
                .put("bytes",source.getBytes(StandardCharsets.UTF_8).length)
                .put("words",source.trim().isEmpty()?0:source.trim().split("\\s+").length)
                .put("lines",source.isEmpty()?0:source.split("\\r\\n|\\n|\\r",-1).length));
            case "url-inspect":return urlInspect(source);
            case "query-parse":return pretty(queryObject(source));
            case "query-stringify":return queryString(source);
            case "openapi-operations":return pretty(operations(source));
            case "openapi-search":return pretty(rank(operations(source),option));
            case "openapi-plan":return pretty(plan(source,option));
            default:throw new IllegalArgumentException("Unknown operation: "+id);
        }
    }
}