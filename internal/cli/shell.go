package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/bspippi1337/restless/internal/httpx"
	"github.com/bspippi1337/restless/internal/store"
	"github.com/bspippi1337/restless/internal/util"
)

func NewShellCmd() *cobra.Command {

	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Interactive API shell (learned endpoints become commands)",
		RunE: func(cmd *cobra.Command, args []string) error {

			cacheRoot, _ := cmd.Root().PersistentFlags().GetString("cache")
			apiName, _ := cmd.Root().PersistentFlags().GetString("api")
			cacheRoot, _ = store.DefaultRoot(cacheRoot)

			api, err := store.Read(cacheRoot, apiName)
			if err != nil {
				return err
			}
			if api == nil || api.BaseURL == "" {
				return fmt.Errorf("no API loaded. Run: restless learn <url>")
			}
			if len(api.Endpoints) == 0 {
				return fmt.Errorf("no endpoints discovered. Run: restless learn <url>")
			}

			epByName := map[string]string{}
			names := make([]string, 0, len(api.Endpoints))

			for _, e := range api.Endpoints {
				n := endpointName(e.Path)
				if n == "" {
					continue
				}
				if _, exists := epByName[n]; !exists {
					epByName[n] = e.Path
					names = append(names, n)
				}
			}
			sort.Strings(names)

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "RESTLESS SHELL\n")
			fmt.Fprintf(out, "Base: %s\n", api.BaseURL)
			fmt.Fprintf(out, "Type: help\n\n")

			client := httpx.New()
			in := bufio.NewScanner(os.Stdin)

			for {
				fmt.Fprint(out, "> ")
				if !in.Scan() {
					fmt.Fprintln(out)
					return nil
				}
				line := strings.TrimSpace(in.Text())
				if line == "" {
					continue
				}

				parts := splitArgs(line)
				head := parts[0]
				tail := parts[1:]

				switch head {
				case "exit", "quit":
					return nil
				case "help":
					printShellHelp(out, names)
					continue
				case "endpoints", "eps":
					for _, n := range names {
						fmt.Fprintf(out, "- %-18s %s\n", n, epByName[n])
					}
					continue
				case "base":
					fmt.Fprintln(out, api.BaseURL)
					continue
				case "call":
					// passthrough: call GET /path [seg...]
					if len(tail) < 2 {
						fmt.Fprintln(out, "usage: call <METHOD> <PATH> [seg...]")
						continue
					}
					method := strings.ToUpper(tail[0])
					path := tail[1]
					if len(tail) > 2 {
						path = appendPath(path, tail[2:])
					}
					if err := doRequest(out, client, api.BaseURL, method, path, timeout); err != nil {
						fmt.Fprintln(out, "error:", err)
					}
					continue
				default:
					path, ok := epByName[head]
					if !ok {
						fmt.Fprintln(out, "unknown command:", head)
						fmt.Fprintln(out, "type: help")
						continue
					}
					if len(tail) > 0 {
						path = appendPath(path, tail)
					}
					if err := doRequest(out, client, api.BaseURL, "GET", path, timeout); err != nil {
						fmt.Fprintln(out, "error:", err)
					}
					continue
				}
			}
		},
	}

	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 12*time.Second, "request timeout")
	return cmd
}

func doRequest(out io.Writer, client *httpx.Client, base, method, path string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	url := util.JoinURL(base, path)

	res, err := client.Do(ctx, method, url, nil)
	if err != nil {
		return err
	}

	body, err := httpx.ReadBody(res, 10<<20)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "%s %s\n", method, url)
	fmt.Fprintf(out, "%s\n", res.Status)
	printMaybeJSON(out, body)
	fmt.Fprintln(out)
	return nil
}

func printMaybeJSON(out io.Writer, body []byte) {
	var v any
	if json.Unmarshal(body, &v) != nil {
		fmt.Fprintln(out, string(body))
		return
	}
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Fprintln(out, string(b))
}

func endpointName(path string) string {
	p := strings.Trim(path, " \t\r\n")
	p = strings.Trim(p, "/")
	if p == "" {
		return ""
	}
	p = strings.ReplaceAll(p, "/", "_")
	p = strings.ReplaceAll(p, "-", "_")
	return p
}

func appendPath(path string, seg []string) string {
	p := strings.TrimRight(path, "/")
	for _, s := range seg {
		s = strings.Trim(s, "/")
		if s == "" {
			continue
		}
		p = p + "/" + s
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

func splitArgs(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	var out []string
	var cur strings.Builder
	inQuote := byte(0)

	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}

	for i := 0; i < len(line); i++ {
		c := line[i]

		if inQuote != 0 {
			if c == inQuote {
				inQuote = 0
				continue
			}
			if c == '\\' && i+1 < len(line) {
				i++
				cur.WriteByte(line[i])
				continue
			}
			cur.WriteByte(c)
			continue
		}

		switch c {
		case ' ', '\t':
			flush()
		case '"', '\'':
			inQuote = c
		default:
			cur.WriteByte(c)
		}
	}
	flush()
	return out
}

func printShellHelp(out io.Writer, names []string) {
	fmt.Fprintln(out, "Commands:")
	fmt.Fprintln(out, "  help                 show this help")
	fmt.Fprintln(out, "  endpoints | eps       list learned endpoint commands")
	fmt.Fprintln(out, "  base                 print base URL")
	fmt.Fprintln(out, "  call METHOD PATH ... raw call (example: call GET /users mojombo)")
	fmt.Fprintln(out, "  exit | quit          leave shell")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Endpoint commands:")
	if len(names) == 0 {
		fmt.Fprintln(out, "  (none)")
		return
	}
	for _, n := range names {
		fmt.Fprintf(out, "  %s [seg...]          GET on learned endpoint (seg appended to path)\n", n)
	}
}
