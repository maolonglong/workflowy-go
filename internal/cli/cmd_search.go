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
		Short: "Search nodes with Workflowy-style query syntax",
		Long: `Search exported Workflowy nodes using a practical subset of Workflowy search syntax.

Supported operators:
  word1 word2         implicit AND
  word1 OR word2      alternatives
  -term               exclude matches
  "exact phrase"      exact phrase
  ancestor > child    nested ancestor search
  is:todo             item type/layout filters
  is:complete         completion filter
  has:note            note presence filter
  created:7d          created within the last 7 days
  changed:24h         changed within the last 24 hours

Unsupported web-only operators such as text:, highlight:, shares, mirrors,
attachments, and backlinks return explicit errors because that data is not
available in Workflowy's export API.`,
		Example: strings.Join([]string{
			`  wf search "project alpha"`,
			`  wf search '#project > is:todo -is:complete'`,
			`  wf search 'meeting > "follow up"'`,
			`  wf search 'has:note OR is:code-block'`,
			`  wf search 'changed:7d is:todo'`,
		}, "\n"),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query, err := compileSearchQuery(args[0])
			if err != nil {
				return err
			}
			nodes, err := a.loadOrExport(cmd.Context(), refresh)
			if err != nil {
				return err
			}
			matches := query.Filter(nodes)

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

	if data, err := json.Marshal(nodes); err == nil {
		_ = writeFileAtomic(cachePath, data, 0o600)
	}

	return nodes, nil
}
