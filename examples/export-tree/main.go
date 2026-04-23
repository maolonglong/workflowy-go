package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/maolonglong/workflowy-go/pkg/workflowy"
)

func main() {
	ctx := context.Background()

	client, err := workflowy.NewClient(
		workflowy.WithAPIKey(os.Getenv("WORKFLOWY_API_KEY")),
	)
	if err != nil {
		log.Fatal(err)
	}

	nodes, err := client.Nodes.Export().Do(ctx)
	if err != nil {
		log.Fatal(err)
	}

	tree := workflowy.BuildTree(nodes)
	fmt.Printf("Top-level nodes: %d\n", len(tree))
	printTree(tree, 0)
}

func printTree(trees []*workflowy.NodeTree, depth int) {
	for _, t := range trees {
		fmt.Printf("%s- %s\n", strings.Repeat("  ", depth), t.Name)
		if len(t.Children) > 0 {
			printTree(t.Children, depth+1)
		}
	}
}
