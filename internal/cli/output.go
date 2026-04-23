package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/maolonglong/workflowy-go/pkg/workflowy"
)

type response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Type    string `json:"type,omitempty"`
}

func printJSON(w io.Writer, resp response) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(resp)
}

func printSuccess(w io.Writer, data any) {
	printJSON(w, response{Success: true, Data: data})
}

func printError(w io.Writer, err error, errType string) {
	printJSON(w, response{Success: false, Error: err.Error(), Type: errType})
}

func truncateOutput(s string, max int) string {
	runes := []rune(s)
	if max <= 0 || len(runes) <= max {
		return s
	}
	return string(runes[:max]) + fmt.Sprintf("\n[truncated: showing %d of %d chars]", max, len(runes))
}

// formatNode formats a node as a single line for human output.
func formatNode(n *workflowy.Node) string {
	status := " "
	if n.CompletedAt != nil {
		status = "✓"
	}
	return fmt.Sprintf("%s  %s  %s", n.ID, status, n.Name)
}

// renderTree writes a node tree with indentation to w.
func renderTree(trees []*workflowy.NodeTree, w io.Writer, depth, maxDepth int) {
	if maxDepth > 0 && depth >= maxDepth {
		return
	}
	for _, t := range trees {
		indent := strings.Repeat("  ", depth)
		status := " "
		if t.CompletedAt != nil {
			status = "✓"
		}
		fmt.Fprintf(w, "%s%s %s %s\n", indent, t.ID, status, t.Name)
		if len(t.Children) > 0 {
			renderTree(t.Children, w, depth+1, maxDepth)
		}
	}
}
