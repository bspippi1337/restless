package cli

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/bspippi1337/restless/internal/engine"
)

func NewGraphCmd() *cobra.Command {

	var format string

	cmd := &cobra.Command{
		Use:   "graph <target>",
		Short: "Generate visual API graph",
		Args:  cobra.ExactArgs(1),

		Run: func(cmd *cobra.Command, args []string) {

			target := args[0]

			res, err := engine.Run(target)
			if err != nil {
				log.Fatal(err)
			}

			dot := engine.TopologyToDOT(res.Topology)

			if format == "dot" {
				fmt.Print(dot)
				return
			}

			out := target + ".svg"

			d := exec.Command("dot", "-Tsvg", "-o", out)
			stdin, _ := d.StdinPipe()

			go func() {
				defer stdin.Close()
				stdin.Write([]byte(dot))
			}()

			if err := d.Run(); err != nil {
				log.Fatal("graphviz 'dot' not found")
			}

			fmt.Println("graph written to", out)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "svg", "output format: svg|dot")

	return cmd
}
