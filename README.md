# workflowy-go

A Go client library for the [WorkFlowy API](https://beta.workflowy.com/api-reference/).

## Installation

```bash
go get github.com/maolonglong/workflowy-go/pkg/workflowy
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/maolonglong/workflowy-go/pkg/workflowy"
)

func main() {
	ctx := context.Background()

	client, err := workflowy.NewClient(
		workflowy.WithAPIKey("wf_your_api_key"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create a node in inbox.
	created, err := client.Nodes.Create("Hello API").
		ParentID(workflowy.TargetInbox).
		Position(workflowy.PositionTop).
		Do(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created:", created.ItemID)

	// Get a node.
	node, err := client.Nodes.Get(created.ItemID).Do(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Name:", node.Name)

	// List children of inbox.
	nodes, err := client.Nodes.List().
		ParentID(workflowy.TargetInbox).
		Do(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, n := range nodes {
		fmt.Printf("- %s\n", n.Name)
	}

	// Update a node.
	err = client.Nodes.Update(created.ItemID).
		Name("Updated title").
		Note("Some notes").
		Do(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Move a node.
	err = client.Nodes.Move(created.ItemID).
		ParentID(workflowy.TargetHome).
		Position(workflowy.PositionBottom).
		Do(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Complete / Uncomplete.
	_ = client.Nodes.Complete(created.ItemID).Do(ctx)
	_ = client.Nodes.Uncomplete(created.ItemID).Do(ctx)

	// Delete.
	_ = client.Nodes.Delete(created.ItemID).Do(ctx)

	// Export all nodes and build tree.
	allNodes, err := client.Nodes.Export().Do(ctx)
	if err != nil {
		log.Fatal(err)
	}
	tree := workflowy.BuildTree(allNodes)
	fmt.Printf("Top-level nodes: %d\n", len(tree))

	// List targets.
	targets, err := client.Targets.List().Do(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, t := range targets {
		fmt.Printf("Target: %s (%s)\n", t.Key, t.Type)
	}
}
```

## CLI

A companion CLI tool `wf` is also included.

### Install

```bash
go install github.com/maolonglong/workflowy-go/cmd/wf@latest
```

### Authenticate

```bash
wf auth login
# Or pipe from stdin:
echo "wf_your_api_key" | wf auth login
# Or use an environment variable:
export WF_API_KEY=wf_your_api_key
```

### Usage

```bash
wf create "Buy groceries" --parent inbox --position top
wf list --parent inbox
wf get <node-id>
wf update <node-id> --name "New title" --note "Some notes"
wf complete <node-id>
wf move <node-id> --parent home
wf delete <node-id>
wf tree --depth 3
wf search "keyword"
wf targets
```

Use `--json` for JSON output (useful for scripting and AI agents):

```bash
wf list --parent inbox --json
```

`wf search` supports a Workflowy-style subset over exported node data:

```bash
wf search 'project alpha'                 # implicit AND
wf search '@me OR @you'
wf search '-#done'
wf search '"exact phrase"'
wf search '#project > is:todo -is:complete'
wf search 'has:note OR is:code-block'
wf search 'created:7d'
wf search 'changed:24h is:todo'
```

Supported filters currently include:

- text and tag matches across node names + notes
- `OR`, unary `-`, and quoted phrases
- nested ancestor search with `>`
- `is:todo`, `is:complete`, `is:bullets`, `is:h1`, `is:h2`, `is:h3`, `is:code-block`, `is:quote-block`
- `has:note`
- `created:<age>` and `changed:<age>` where `<age>` is relative time like `30m`, `12h`, `7d`, or `2w`

Workflowy web operators that require richer export metadata (for example `text:`,
`highlight:`, attachments, mirrors, backlinks, or sharing state) return explicit
errors in the CLI.

## Error Handling

```go
node, err := client.Nodes.Get("non-existent-id").Do(ctx)
if workflowy.IsNotFound(err) {
	fmt.Println("Node not found")
}
if workflowy.IsRateLimited(err) {
	fmt.Println("Rate limited, try again later")
}
```

## License

MIT
