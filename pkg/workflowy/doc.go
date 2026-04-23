// Package workflowy provides a Go client for the WorkFlowy API.
//
// Create a client with an API key and use it to manage nodes and targets:
//
//	client, err := workflowy.NewClient(workflowy.WithAPIKey("wf_your_api_key"))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create a node.
//	created, err := client.Nodes.Create("Hello API").
//		ParentID(workflowy.TargetInbox).
//		Do(ctx)
//
// See https://beta.workflowy.com/api-reference/ for API documentation.
package workflowy
