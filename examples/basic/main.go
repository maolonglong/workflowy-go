package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

	// List targets.
	targets, err := client.Targets.List().Do(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, t := range targets {
		fmt.Printf("Target: %s (%s)\n", t.Key, t.Type)
	}

	// Create a node in inbox.
	created, err := client.Nodes.Create("Hello API").
		ParentID(workflowy.TargetInbox).
		Position(workflowy.PositionTop).
		Do(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created node: %s\n", created.ItemID)

	// Get the node.
	node, err := client.Nodes.Get(created.ItemID).Do(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Node: %s (priority=%g)\n", node.Name, node.Priority)

	// Update the node.
	err = client.Nodes.Update(created.ItemID).
		Name("Updated title").
		Do(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Complete the node.
	err = client.Nodes.Complete(created.ItemID).Do(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Delete the node.
	err = client.Nodes.Delete(created.ItemID).Do(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Node deleted.")
}
