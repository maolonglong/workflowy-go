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
			id, err := a.resolveNodeID(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if err := a.client.Nodes.Complete(id).Do(cmd.Context()); err != nil {
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
			id, err := a.resolveNodeID(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if err := a.client.Nodes.Uncomplete(id).Do(cmd.Context()); err != nil {
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
			id, err := a.resolveNodeID(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			call := a.client.Nodes.Move(id)
			if parent != "" {
				parentID, err := a.resolveNodeID(cmd.Context(), parent)
				if err != nil {
					return err
				}
				call = call.ParentID(workflowy.ParentRef(parentID))
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
