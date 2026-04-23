package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/maolonglong/workflowy-go/pkg/workflowy"
)

func (a *app) newCompleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "complete <id>",
		Short: "Mark a node as completed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.client.Nodes.Complete(args[0]).Do(cmd.Context()); err != nil {
				return err
			}
			if a.jsonOutput {
				printSuccess(os.Stdout, map[string]string{"status": "ok"})
			} else {
				fmt.Println("Completed.")
			}
			return nil
		},
	}
}

func (a *app) newUncompleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uncomplete <id>",
		Short: "Mark a node as not completed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.client.Nodes.Uncomplete(args[0]).Do(cmd.Context()); err != nil {
				return err
			}
			if a.jsonOutput {
				printSuccess(os.Stdout, map[string]string{"status": "ok"})
			} else {
				fmt.Println("Uncompleted.")
			}
			return nil
		},
	}
}

func (a *app) newMoveCmd() *cobra.Command {
	var parent, position string

	cmd := &cobra.Command{
		Use:   "move <id>",
		Short: "Move a node to a new parent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			call := a.client.Nodes.Move(args[0])
			if parent != "" {
				call = call.ParentID(workflowy.ParentRef(parent))
			}
			if position != "" {
				call = call.Position(workflowy.Position(position))
			}
			if err := call.Do(cmd.Context()); err != nil {
				return err
			}
			if a.jsonOutput {
				printSuccess(os.Stdout, map[string]string{"status": "ok"})
			} else {
				fmt.Println("Moved.")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&parent, "parent", "", "new parent (UUID or target key)")
	cmd.Flags().StringVar(&position, "position", "", "position (top or bottom)")

	return cmd
}
