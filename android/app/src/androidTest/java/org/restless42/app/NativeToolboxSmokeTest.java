package org.restless42.app;

import android.view.View;
import android.view.ViewGroup;
import android.webkit.WebView;
import android.widget.TextView;

import androidx.test.core.app.ActivityScenario;
import androidx.test.ext.junit.runners.AndroidJUnit4;

import org.json.JSONArray;
import org.json.JSONObject;
import org.junit.Test;
import org.junit.runner.RunWith;

import java.util.ArrayList;
import java.util.List;

import static org.junit.Assert.*;

/** Exercises real Android JSON and native Android widgets on an emulator. */
@RunWith(AndroidJUnit4.class)
public class NativeToolboxSmokeTest {
    @Test public void all32OperationsRegistered(){
        assertEquals(32,ToolboxEngine.TOOLS.size());
        java.util.Set<String> names=new java.util.HashSet<>();
        for(ToolboxEngine.Tool tool:ToolboxEngine.TOOLS)assertTrue("duplicate "+tool.id,names.add(tool.id));
    }

    @Test public void utf8TransformsAreReversible() throws Exception {
        String input="Hei 🌍! RESTLESS 42";
        assertEquals(input,ToolboxEngine.apply("base64-decode",ToolboxEngine.apply("base64-encode",input,""),""));
        assertEquals(input,ToolboxEngine.apply("base64url-decode",ToolboxEngine.apply("base64url-encode",input,""),""));
        assertEquals(input,ToolboxEngine.apply("hex-decode",ToolboxEngine.apply("hex-encode",input,""),""));
    }

    @Test public void jsonCsvAndRecipesActuallyRun() throws Exception {
        String json="[{\"name\":\"Alex, Jr.\",\"status\":\"OK\"},{\"name\":\"Sam\",\"status\":\"RUNNING\"}]";
        String csv=ToolboxEngine.apply("json-to-csv",json,"");
        assertTrue(csv.contains("\"Alex, Jr.\""));
        JSONArray roundTrip=new JSONArray(ToolboxEngine.apply("csv-to-json",csv,""));
        assertEquals("Sam",roundTrip.getJSONObject(1).getString("name"));
        List<ToolboxEngine.Step> steps=new ArrayList<>();
        steps.add(new ToolboxEngine.Step("json-pretty","",true));
        steps.add(new ToolboxEngine.Step("base64-encode","",true));
        String result=ToolboxEngine.run("{\"demo\":42}",steps);
        assertTrue(result.length()>10);
        List<ToolboxEngine.Step> imported=ToolboxEngine.importRecipe(ToolboxEngine.exportRecipe(steps).toString());
        assertEquals(2,imported.size());
        assertEquals(result,ToolboxEngine.run("{\"demo\":42}",imported));
    }

    @Test public void sha256MatchesKnownVector() throws Exception {
        assertEquals("ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
                ToolboxEngine.apply("sha-256","abc",""));
    }

    @Test public void jwtInspectionNeverClaimsSignatureTrust() throws Exception {
        String jwt="eyJhbGciOiJub25lIn0.eyJzdWIiOiJkZW1vIn0.";
        JSONObject output=new JSONObject(ToolboxEngine.apply("jwt-inspect",jwt,""));
        assertFalse(output.getBoolean("signatureVerified"));
        assertEquals("demo",output.getJSONObject("payload").getString("sub"));
    }

    @Test public void offlinePlanHasZeroNetworkRequests() throws Exception {
        String spec="{\"openapi\":\"3.0.3\",\"info\":{\"title\":\"Pets\",\"version\":\"1\"},\"paths\":{\"/pets/{petId}\":{\"get\":{\"summary\":\"Find pet by ID\",\"operationId\":\"findPetById\",\"parameters\":[{\"in\":\"path\",\"name\":\"petId\",\"required\":true}]}}}}";
        JSONObject plan=new JSONObject(ToolboxEngine.apply("openapi-plan",spec,"find pet by id"));
        assertEquals(0,plan.getInt("networkRequests"));
        assertEquals("DOCUMENTATION_ONLY_NO_EXECUTION",plan.getString("mode"));
        JSONArray candidates=plan.getJSONArray("candidates");
        assertTrue(candidates.length()>0);
        assertTrue(candidates.getJSONObject(0).getBoolean("reviewRequired"));
    }

    @Test public void activityLaunchesNativeViewsWithNoWebView(){
        try(ActivityScenario<MainActivity> scenario=ActivityScenario.launch(MainActivity.class)){
            scenario.onActivity(activity->{
                View root=activity.getWindow().getDecorView();
                assertTrue("Native workbench header is missing",containsText(root,"RESTLESS 42"));
                assertTrue("Expected native workbench",containsText(root,"An entire API workbench."));
                assertTrue("Expected native offline badge",containsText(root,"NO NETWORK"));
                assertEquals("WebView is not permitted",0,countWebViews(root));
            });
        }
    }

    private static boolean containsText(View view,String text){
        if(view instanceof TextView&&((TextView)view).getText().toString().contains(text))return true;
        if(view instanceof ViewGroup){
            ViewGroup group=(ViewGroup)view;
            for(int i=0;i<group.getChildCount();i++)if(containsText(group.getChildAt(i),text))return true;
        }
        return false;
    }
    private static int countWebViews(View view){
        int count=view instanceof WebView?1:0;
        if(view instanceof ViewGroup){
            ViewGroup group=(ViewGroup)view;
            for(int i=0;i<group.getChildCount();i++)count+=countWebViews(group.getChildAt(i));
        }
        return count;
    }
}
