package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/maolonglong/workflowy-go/pkg/workflowy"
)

func (a *app) newIDCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "id <id>",
		Short: "Resolve a node ID, short ID, internal link, or target key to a UUID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := a.resolveNodeUUID(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if a.jsonOutput {
				printSuccess(os.Stdout, map[string]string{"id": id})
			} else {
				fmt.Println(id)
			}
			return nil
		},
	}
}

// resolveNodeID resolves user input to a value safe to pass to the WorkFlowy
// API as a node identifier. Full UUIDs and short IDs are normalized to the
// full UUID; everything else (target keys like "inbox") is passed through
// unchanged so the API can resolve it.
func (a *app) resolveNodeID(ctx context.Context, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == string(workflowy.ParentNone) {
		return raw, nil
	}

	sanitized := workflowy.SanitizeNodeID(raw)
	switch {
	case workflowy.IsNodeUUID(sanitized):
		return sanitized, nil
	case workflowy.IsShortID(sanitized):
		nodes, err := a.loadOrExport(ctx, false)
		if err != nil {
			return "", err
		}
		return workflowy.ResolveShortIDFromNodes(nodes, sanitized)
	default:
		// Treat as target key; the API will reject unknown keys.
		return raw, nil
	}
}

// resolveNodeUUID is like resolveNodeID but always returns a full UUID,
// dereferencing target keys via the API.
func (a *app) resolveNodeUUID(ctx context.Context, raw string) (string, error) {
	id, err := a.resolveNodeID(ctx, raw)
	if err != nil {
		return "", err
	}
	if id == "" || id == string(workflowy.ParentNone) {
		return id, nil
	}
	if workflowy.IsNodeUUID(id) {
		return id, nil
	}
	node, err := a.client.Nodes.Get(id).Do(ctx)
	if err != nil {
		return "", fmt.Errorf("workflowy: resolve %q: %w", raw, err)
	}
	return node.ID, nil
}
