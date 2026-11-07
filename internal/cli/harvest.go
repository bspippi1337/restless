package cli

import (
    "fmt"
    "time"

    "github.com/bspippi1337/restless/internal/botanical"
    "github.com/spf13/cobra"
)

func NewHarvestCmd() *cobra.Command {
    var host, out, agent string
    var pages, depth int
    var delay time.Duration
    var execute, authorized bool

    cmd := &cobra.Command{
        Use: "harvest <https-url>",
        Short: "Scope-locked, robots-aware API documentation harvester (dry-run by default)",
        Long: "Archive approved public API docs and white papers with SHA-256 provenance, respecting robots.txt and a strict page budget. Only crawl systems you are authorized to access.",
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            if host == "" { return fmt.Errorf("--host is required as an explicit crawl allowlist") }
            if !execute {
                _, _ = fmt.Fprintf(cmd.OutOrStdout(),
                    "DRY RUN: zero network requests\nSeed: %s\nAllowed host: %s\nMax pages: %d\nMax depth: %d\nDelay: %s\nArchive: %s\nTo run: add --execute --authorized (only after checking site terms and permission)\n",
                    args[0], host, pages, depth, delay, out)
                return nil
            }
            records, err := botanical.Crawl(cmd.Context(), botanical.Config{
                SeedURL: args[0], AllowedHost: host, Agent: agent,
                OutputDir: out, MaxPages: pages, MaxDepth: depth,
                Delay: delay, Authorized: authorized,
            })
            for _, record := range records {
                _, _ = fmt.Fprintf(cmd.OutOrStdout(), "archived %s sha256=%s bytes=%d\n",
                    record.URL, record.SHA256, record.Bytes)
            }
            if err != nil { return err }
            _, _ = fmt.Fprintf(cmd.OutOrStdout(), "Complete: %d documents indexed in %s/index.jsonl\n", len(records), out)
            return nil
        },
    }
    cmd.Flags().StringVar(&host, "host", "", "exact permitted hostname (required)")
    cmd.Flags().StringVar(&out, "out", ".restless/botanical", "local documentation archive path")
    cmd.Flags().StringVar(&agent, "agent", "RestlessBot/42 (+https://github.com/bspippi1337/restless)", "identifiable crawler User-Agent")
    cmd.Flags().IntVar(&pages, "max-pages", 5, "maximum archived pages (1..25)")
    cmd.Flags().IntVar(&depth, "max-depth", 1, "maximum link depth (0..3)")
    cmd.Flags().DurationVar(&delay, "delay", 2*time.Second, "minimum interval between page requests (>=2s)")
    cmd.Flags().BoolVar(&execute, "execute", false, "enable network access")
    cmd.Flags().BoolVar(&authorized, "authorized", false, "confirm permission and compliance with target terms")
    return cmd
}
