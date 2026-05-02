package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/maolonglong/workflowy-go/pkg/workflowy"
)

func (a *app) newCreateCmd() *cobra.Command {
	var parent, note, layout, position string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			call := a.client.Nodes.Create(args[0])
			if parent != "" {
				parentID, err := a.resolveNodeID(cmd.Context(), parent)
				if err != nil {
					return err
				}
				call = call.ParentID(workflowy.ParentRef(parentID))
			}
			if note != "" {
				call = call.Note(note)
			}
			if layout != "" {
				call = call.Layout(workflowy.LayoutMode(layout))
			}
			if position != "" {
				call = call.Position(workflowy.Position(position))
			}
			resp, err := call.Do(cmd.Context())
			if err != nil {
				return err
			}
			if a.jsonOutput {
				printSuccess(os.Stdout, resp)
			} else {
				fmt.Printf("Created: %s\n", resp.ItemID)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&parent, "parent", "", "parent (UUID or target key)")
	cmd.Flags().StringVar(&note, "note", "", "note content")
	cmd.Flags().StringVar(&layout, "layout", "", "layout mode (bullets, todo, h1, h2, h3, code-block, quote-block)")
	cmd.Flags().StringVar(&position, "position", "", "position (top or bottom)")

	return cmd
}

func (a *app) newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := a.resolveNodeID(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			node, err := a.client.Nodes.Get(id).Do(cmd.Context())
			if err != nil {
				return err
			}
			if a.jsonOutput {
				printSuccess(os.Stdout, node)
			} else {
				fmt.Println(formatNode(node))
				if node.Note != nil && *node.Note != "" {
					fmt.Printf("  Note: %s\n", *node.Note)
				}
				fmt.Printf("  Layout: %s\n", node.Data.LayoutMode)
				fmt.Printf("  Created: %s\n", node.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("  Modified: %s\n", node.ModifiedAt.Format("2006-01-02 15:04:05"))
			}
			return nil
		},
	}
}

func (a *app) newListCmd() *cobra.Command {
	var parent string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List child nodes",
		RunE: func(cmd *cobra.Command, args []string) error {
			call := a.client.Nodes.List()
			if parent != "" {
				parentID, err := a.resolveNodeID(cmd.Context(), parent)
				if err != nil {
					return err
				}
				call = call.ParentID(workflowy.ParentRef(parentID))
			}
			nodes, err := call.Do(cmd.Context())
			if err != nil {
				return err
			}
			if a.jsonOutput {
				printSuccess(os.Stdout, nodes)
			} else {
				for _, n := range nodes {
					fmt.Println(formatNode(n))
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&parent, "parent", "", "parent (UUID or target key)")
	return cmd
}

func (a *app) newUpdateCmd() *cobra.Command {
	var name, note, layout string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := a.resolveNodeID(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			call := a.client.Nodes.Update(id)
			if cmd.Flags().Changed("name") {
				call = call.Name(name)
			}
			if cmd.Flags().Changed("note") {
				call = call.Note(note)
			}
			if cmd.Flags().Changed("layout") {
				call = call.Layout(workflowy.LayoutMode(layout))
			}
			if err := call.Do(cmd.Context()); err != nil {
				return err
			}
			if a.jsonOutput {
				printSuccess(os.Stdout, map[string]string{"status": "ok"})
			} else {
				fmt.Println("Updated.")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "new name")
	cmd.Flags().StringVar(&note, "note", "", "new note")
	cmd.Flags().StringVar(&layout, "layout", "", "new layout mode")

	return cmd
}

func (a *app) newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a node permanently",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := a.resolveNodeID(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if err := a.client.Nodes.Delete(id).Do(cmd.Context()); err != nil {
				return err
			}
			if a.jsonOutput {
				printSuccess(os.Stdout, map[string]string{"status": "ok"})
			} else {
				fmt.Println("Deleted.")
			}
			return nil
		},
	}
}
