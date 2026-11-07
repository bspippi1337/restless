package cli

import (
	"encoding/json"
	"fmt"

	"github.com/bspippi1337/restless/internal/botanical"
	"github.com/spf13/cobra"
)

func NewIndexCmd() *cobra.Command {
	var dir string
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Verify archived documents and build a local OpenAPI catalog",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := botanical.BuildCatalog(dir)
			if err != nil {
				return err
			}
			if err := botanical.SaveCatalog(dir, c); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Indexed %d documents and %d documented operations in %s/catalog.json\n", len(c.Documents), len(c.Operations), dir)
			return err
		},
	}
	cmd.Flags().StringVar(&dir, "dir", ".restless/botanical", "local harvester archive directory")
	return cmd
}

func NewSearchCmd() *cobra.Command {
	var dir string
	var limit int
	cmd := &cobra.Command{
		Use:   "search <keywords>",
		Short: "Rank documented API operations locally without requests",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := botanical.LoadCatalog(dir)
			if err != nil {
				return err
			}
			hits, err := botanical.RankOperations(c, args[0], limit)
			if err != nil {
				return err
			}
			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", "  ")
			return encoder.Encode(hits)
		},
	}
	cmd.Flags().StringVar(&dir, "dir", ".restless/botanical", "catalog directory")
	cmd.Flags().IntVar(&limit, "limit", 10, "max returned operations (1..50)")
	return cmd
}

func NewPlanCmd() *cobra.Command {
	var dir string
	var limit int
	cmd := &cobra.Command{
		Use:   "plan <task>",
		Short: "Produce a provenance-backed API probe plan without executing calls",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := botanical.LoadCatalog(dir)
			if err != nil {
				return err
			}
			plan, err := botanical.BuildProbePlan(c, args[0], limit)
			if err != nil {
				return err
			}
			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", "  ")
			return encoder.Encode(plan)
		},
	}
	cmd.Flags().StringVar(&dir, "dir", ".restless/botanical", "catalog directory")
	cmd.Flags().IntVar(&limit, "limit", 5, "max plan candidates (1..50)")
	return cmd
}

func NewRefsCmd() *cobra.Command {
 var dir string
 var limit int
 cmd:=&cobra.Command{
  Use:"refs <keywords>",
  Short:"Search locally archived reference documents and whitepapers",
  Args:cobra.ExactArgs(1),
  RunE:func(cmd *cobra.Command,args []string)error{
   c,err:=botanical.LoadCatalog(dir)
   if err!=nil{return err}
   results,err:=botanical.RankReferences(c,args[0],limit)
   if err!=nil{return err}
   encoder:=json.NewEncoder(cmd.OutOrStdout())
   encoder.SetIndent("","  ")
   return encoder.Encode(results)
  },
 }
 cmd.Flags().StringVar(&dir,"dir",".restless/botanical","catalog directory")
 cmd.Flags().IntVar(&limit,"limit",10,"max reference results (1..50)")
 return cmd
}
