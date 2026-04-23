package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/maolonglong/workflowy-go/pkg/workflowy"
)

func (a *app) newTreeCmd() *cobra.Command {
	var maxDepth int
	var refresh bool

	cmd := &cobra.Command{
		Use:   "tree",
		Short: "Display full node tree",
		RunE: func(cmd *cobra.Command, args []string) error {
			nodes, err := a.loadOrExport(cmd.Context(), refresh)
			if err != nil {
				return err
			}
			trees := workflowy.BuildTree(nodes)

			if a.jsonOutput {
				printSuccess(os.Stdout, trees)
			} else {
				var buf strings.Builder
				renderTree(trees, &buf, 0, maxDepth)
				output := buf.String()
				if a.maxOutput > 0 {
					output = truncateOutput(output, a.maxOutput)
				}
				fmt.Print(output)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&maxDepth, "depth", 0, "max tree depth (0 = unlimited)")
	cmd.Flags().BoolVar(&refresh, "refresh", false, "force refresh from API")

	return cmd
}
