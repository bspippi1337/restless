package cli

import (
 "encoding/json"
 "fmt"
 "time"

 "github.com/bspippi1337/restless/internal/botanical"
 "github.com/bspippi1337/restless/internal/library"
 "github.com/spf13/cobra"
)

// NewLibraryCmd turns one explicitly supplied domain into a source-backed
// documentation and OpenAPI catalog. Preview is the default and never connects.
func NewLibraryCmd()*cobra.Command{
 var out,agent string
 var pages,depth int
 var delay time.Duration
 var execute,authorized,jsonOutput bool
 cmd:=&cobra.Command{
  Use:"library <domain-or-https-docs-url>",
  Short:"Build an offline API library from linked documentation, not guessed endpoints",
  Long:"Preview or perform one authorized, robots-aware, bounded documentation crawl of a single exact hostname. The result is an auditable SHA-256 document library and an offline OpenAPI method/path catalog. No API operations are executed.",
  Args:cobra.ExactArgs(1),
  RunE:func(cmd *cobra.Command,args []string)error{
   opts:=library.Options{
    Input:args[0],OutputDir:out,MaxPages:pages,MaxDepth:depth,
    Delay:delay,Agent:agent,Execute:execute,Authorized:authorized,
   }
   if !jsonOutput&&execute{
    opts.OnDocument=func(r botanical.Record){
     fmt.Fprintf(cmd.OutOrStdout(),"ARCHIVED %s sha256=%s (%d bytes)\n",r.URL,r.SHA256,r.Bytes)
    }
   }
   result,err:=library.Build(cmd.Context(),opts)
   if err!=nil{return err}
   if jsonOutput{
    encoder:=json.NewEncoder(cmd.OutOrStdout());encoder.SetIndent("","  ")
    return encoder.Encode(result)
   }
   if result.Mode=="DRY_RUN"{
    _,err=fmt.Fprintf(cmd.OutOrStdout(),
     "DRY RUN / NO NETWORK\nHost: %s\nSeed: %s\nArchive: %s\nMax documents: %d\nMax link depth: %d\nMin pause: %ds\n\nTo build the library: add --execute --authorized only after checking permission and scope.\n",
     result.Target.Host,result.Target.SeedURL,result.OutputDir,result.MaxPages,result.MaxDepth,result.DelaySeconds)
    return err
   }
   _,err=fmt.Fprintf(cmd.OutOrStdout(),
    "LIBRARY COMPLETE\nDocuments: %d\nDocumented operations: %d\nCatalog: %s/catalog.json\nLibrary: %s/library.json\nNo API operations were executed.\n",
    result.DocumentCount,result.OperationCount,result.OutputDir,result.OutputDir)
   return err
  },
 }
 cmd.Flags().StringVar(&out,"out",".restless/libraries","base directory for one per-domain library")
 cmd.Flags().StringVar(&agent,"agent",library.DefaultAgent,"identifiable crawler user-agent")
 cmd.Flags().IntVar(&pages,"max-pages",12,"maximum documents (1..25)")
 cmd.Flags().IntVar(&depth,"max-depth",2,"maximum documentation link depth (0..3)")
 cmd.Flags().DurationVar(&delay,"delay",2*time.Second,"minimum spacing (>=2s)")
 cmd.Flags().BoolVar(&execute,"execute",false,"opt in to network requests")
 cmd.Flags().BoolVar(&authorized,"authorized",false,"confirm exact-host authorization, policy and site terms")
 cmd.Flags().BoolVar(&jsonOutput,"json",false,"print machine-readable JSON summary")
 return cmd
}
