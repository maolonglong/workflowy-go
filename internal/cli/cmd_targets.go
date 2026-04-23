package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func (a *app) newTargetsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "targets",
		Short: "List available targets (inbox, home, shortcuts)",
		RunE: func(cmd *cobra.Command, args []string) error {
			targets, err := a.client.Targets.List().Do(cmd.Context())
			if err != nil {
				return err
			}
			if a.jsonOutput {
				printSuccess(os.Stdout, targets)
			} else {
				for _, t := range targets {
					name := "(not created)"
					if t.Name != nil {
						name = *t.Name
					}
					fmt.Printf("%-12s %-10s %s\n", t.Key, t.Type, name)
				}
			}
			return nil
		},
	}
}
