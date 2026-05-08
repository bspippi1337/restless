package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/bspippi1337/restless/internal/events"
	"github.com/bspippi1337/restless/internal/observe"
	"github.com/bspippi1337/restless/internal/pipeline"
	"github.com/bspippi1337/restless/internal/watch"
)

func NewWatchCmd() *cobra.Command {
	var command string
	var debounce int
	var jsonMode bool

	cmd := &cobra.Command{
		Use:   "watch <path>",
		Short: "Reactive filesystem runtime",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if command == "" {
				return fmt.Errorf("missing --run command")
			}

			return watch.Run(args[0], time.Duration(debounce)*time.Millisecond, func(ev events.Event) {
				exec := pipeline.Run(ev, command)

				if jsonMode {
					_ = observe.PrintJSON(exec)
					return
				}

				observe.PrintHuman(exec)
			})
		},
	}

	cmd.Flags().StringVarP(&command, "run", "r", "", "command to execute")
	cmd.Flags().IntVar(&debounce, "debounce", 250, "debounce in milliseconds")
	cmd.Flags().BoolVar(&jsonMode, "json", false, "emit JSON runtime events")

	return cmd
}
