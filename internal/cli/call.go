package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/bspippi1337/restless/internal/httpx"
	"github.com/bspippi1337/restless/internal/store"
	"github.com/bspippi1337/restless/internal/util"
	"github.com/spf13/cobra"
)

func NewCallCmd() *cobra.Command {

	var timeout time.Duration
	var table bool

	cmd := &cobra.Command{
		Use:   "call <METHOD> <PATH>",
		Short: "Call API endpoint",
		Args:  cobra.ExactArgs(2),

		RunE: func(cmd *cobra.Command, args []string) error {

			method := strings.ToUpper(args[0])
			path := args[1]

			apiName, _ := cmd.Root().PersistentFlags().GetString("api")
			cacheRoot, _ := cmd.Root().PersistentFlags().GetString("cache")

			cacheRoot, _ = store.DefaultRoot(cacheRoot)

			api, err := store.Read(cacheRoot, apiName)
			if err != nil {
				return err
			}

			url := util.JoinURL(api.BaseURL, path)

			ctx, cancel := context.WithTimeout(context.Background(), timeout)

			defer cancel()
			client := httpx.New()

			res, err := client.Do(ctx, method, url, nil)
			if err != nil {
				return err
			}

			body, err := io.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				return err
			}

			fmt.Println(method, url)
			fmt.Println(res.Status)

			if table {
				return renderTable(body)
			}

			return renderJSON(body)
		},
	}

	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 10*time.Second, "timeout")
	cmd.Flags().BoolVar(&table, "table", false, "render JSON array as table")

	return cmd
}

func renderJSON(body []byte) error {

	var v interface{}

	if json.Unmarshal(body, &v) != nil {
		fmt.Println(string(body))
		return nil
	}

	pretty, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(pretty))
	return nil
}

func renderTable(body []byte) error {

	var rows []map[string]interface{}

	if json.Unmarshal(body, &rows) != nil {
		return renderJSON(body)
	}

	if len(rows) == 0 {
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	var headers []string
	for k := range rows[0] {
		headers = append(headers, k)
	}

	fmt.Fprintln(w, strings.Join(headers, "\t"))

	for _, r := range rows {
		var line []string
		for _, h := range headers {
			line = append(line, fmt.Sprintf("%v", r[h]))
		}
		fmt.Fprintln(w, strings.Join(line, "\t"))
	}

	w.Flush()

	return nil
}
