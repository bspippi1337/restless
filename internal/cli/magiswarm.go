package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bspippi1337/restless/internal/core/magiswarm"
	"github.com/spf13/cobra"
)

func NewMagiswarmCmd() *cobra.Command {

	var concurrency int
	var maxReq int
	var timeout time.Duration
	var outDir string
	var wordlist string
	var noFuzz bool
	var header []string

	cmd := &cobra.Command{
		Use:   "magiswarm <url>",
		Short: "API recon engine: discover, fuzz, map, report",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			target := strings.TrimSpace(args[0])
			opt := magiswarm.DefaultOptions(target)

			if concurrency > 0 {
				opt.Concurrency = concurrency
			}
			if maxReq > 0 {
				opt.MaxRequests = maxReq
			}
			if timeout > 0 {
				opt.Timeout = timeout
			}
			opt.EnableFuzz = !noFuzz

			for _, h := range header {
				parts := strings.SplitN(h, ":", 2)
				if len(parts) != 2 {
					continue
				}
				k := strings.TrimSpace(parts[0])
				v := strings.TrimSpace(parts[1])
				if k != "" && v != "" {
					opt.Headers[k] = v
				}
			}

			if wordlist != "" {
				b, err := os.ReadFile(wordlist)
				if err != nil {
					return err
				}
				opt.Wordlist = strings.Split(string(b), "\n")
			}

			fmt.Println("\033[38;5;45m⚡ restless magiswarm\033[0m")
			fmt.Println("sandstorm pulse: on  |  nagios: shhh")
			fmt.Println()

			r, err := magiswarm.New(opt)
			if err != nil {
				return err
			}

			ctx := context.Background()
			rep, err := r.Run(ctx)
			if err != nil {
				return err
			}

			jsonPath, topPath, err := magiswarm.WriteReportFiles(rep, outDir)
			if err != nil {
				return err
			}

			fmt.Println(rep.Topology)
			fmt.Printf("found: %d unique paths (%d requests, %d errors)\n", rep.Stats.Unique, rep.Stats.Requests, rep.Stats.Errors)
			fmt.Println("report:", jsonPath)
			fmt.Println("topology:", topPath)
			return nil
		},
	}

	cmd.Flags().IntVar(&concurrency, "concurrency", 8, "parallel workers")
	cmd.Flags().IntVar(&maxReq, "max-requests", 200, "max HTTP requests")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Second, "per-request timeout")
	cmd.Flags().StringVar(&outDir, "out", "dist", "output directory for reports")
	cmd.Flags().StringVar(&wordlist, "wordlist", "", "path wordlist file (one per line)")
	cmd.Flags().BoolVar(&noFuzz, "no-fuzz", false, "disable query fuzzing")
	cmd.Flags().StringArrayVar(&header, "header", nil, "extra header (repeatable), e.g. --header 'Authorization: Bearer ...'")

	return cmd
}
