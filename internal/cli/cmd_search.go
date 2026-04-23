package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/maolonglong/workflowy-go/pkg/workflowy"
)

func (a *app) newSearchCmd() *cobra.Command {
	var refresh bool

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search nodes by name or note text",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.ToLower(args[0])
			nodes, err := a.loadOrExport(cmd.Context(), refresh)
			if err != nil {
				return err
			}

			var matches []*workflowy.Node
			for _, n := range nodes {
				if strings.Contains(strings.ToLower(n.Name), query) {
					matches = append(matches, n)
					continue
				}
				if n.Note != nil && strings.Contains(strings.ToLower(*n.Note), query) {
					matches = append(matches, n)
				}
			}

			if a.jsonOutput {
				printSuccess(os.Stdout, matches)
			} else {
				if len(matches) == 0 {
					fmt.Println("No results found.")
					return nil
				}
				for _, n := range matches {
					fmt.Println(formatNode(n))
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&refresh, "refresh", false, "force refresh cache from API")
	return cmd
}

const cacheTTL = 5 * time.Minute

// loadOrExport returns cached export data if fresh enough, otherwise fetches from API.
func (a *app) loadOrExport(ctx context.Context, forceRefresh bool) ([]*workflowy.Node, error) {
	dir, err := cacheDir()
	if err != nil {
		return nil, err
	}
	cachePath := filepath.Join(dir, "export.json")

	if !forceRefresh {
		info, err := os.Stat(cachePath)
		if err == nil && time.Since(info.ModTime()) < cacheTTL {
			data, err := os.ReadFile(cachePath)
			if err == nil {
				var nodes []*workflowy.Node
				if json.Unmarshal(data, &nodes) == nil {
					return nodes, nil
				}
			}
		}
	}

	nodes, err := a.client.Nodes.Export().Do(ctx)
	if err != nil {
		return nil, err
	}

	// Save to cache.
	if err := os.MkdirAll(dir, 0o700); err == nil {
		if data, err := json.Marshal(nodes); err == nil {
			_ = os.WriteFile(cachePath, data, 0o600)
		}
	}

	return nodes, nil
}
