package workflowy

import (
	"context"
	"net/url"
	"sort"
)

// NodesService handles operations on WorkFlowy nodes.
type NodesService struct {
	s *service
}

// Get returns a call that retrieves a single node by ID.
func (ns *NodesService) Get(id string) *NodesGetCall {
	return &NodesGetCall{s: ns.s, id: id}
}

// List returns a call that lists child nodes.
func (ns *NodesService) List() *NodesListCall {
	return &NodesListCall{s: ns.s}
}

// Create returns a call that creates a new node.
func (ns *NodesService) Create(name string) *NodesCreateCall {
	return &NodesCreateCall{
		s:    ns.s,
		body: &createNodeBody{Name: name},
	}
}

// Update returns a call that updates an existing node.
func (ns *NodesService) Update(id string) *NodesUpdateCall {
	return &NodesUpdateCall{
		s:    ns.s,
		id:   id,
		body: &updateNodeBody{},
	}
}

// Delete returns a call that permanently deletes a node.
func (ns *NodesService) Delete(id string) *NodesDeleteCall {
	return &NodesDeleteCall{s: ns.s, id: id}
}

// Move returns a call that moves a node to a new parent.
func (ns *NodesService) Move(id string) *NodesMoveCall {
	return &NodesMoveCall{
		s:    ns.s,
		id:   id,
		body: &moveNodeBody{},
	}
}

// Complete returns a call that marks a node as completed.
func (ns *NodesService) Complete(id string) *NodesCompleteCall {
	return &NodesCompleteCall{s: ns.s, id: id}
}

// Uncomplete returns a call that marks a node as not completed.
func (ns *NodesService) Uncomplete(id string) *NodesUncompleteCall {
	return &NodesUncompleteCall{s: ns.s, id: id}
}

// Export returns a call that exports all nodes as a flat list.
func (ns *NodesService) Export() *NodesExportCall {
	return &NodesExportCall{s: ns.s}
}

// --- Get ---

// NodesGetCall retrieves a single node.
type NodesGetCall struct {
	s  *service
	id string
}

// Do executes the get call.
func (c *NodesGetCall) Do(ctx context.Context) (*Node, error) {
	req, err := c.s.newRequest("GET", "/nodes/"+c.id, nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Node *Node `json:"node"`
	}
	if err := c.s.do(ctx, req, &resp); err != nil {
		return nil, err
	}
	return resp.Node, nil
}

// --- List ---

// NodesListCall lists child nodes of a parent.
type NodesListCall struct {
	s        *service
	parentID ParentRef
}

// ParentID sets the parent to list children of.
func (c *NodesListCall) ParentID(p ParentRef) *NodesListCall {
	c.parentID = p
	return c
}

// Do executes the list call. Results are sorted by priority ascending.
func (c *NodesListCall) Do(ctx context.Context) ([]*Node, error) {
	params := url.Values{}
	if c.parentID != "" {
		params.Set("parent_id", string(c.parentID))
	}
	path := "/nodes"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}
	req, err := c.s.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Nodes []*Node `json:"nodes"`
	}
	if err := c.s.do(ctx, req, &resp); err != nil {
		return nil, err
	}
	sort.Slice(resp.Nodes, func(i, j int) bool {
		return resp.Nodes[i].Priority < resp.Nodes[j].Priority
	})
	return resp.Nodes, nil
}

// --- Create ---

type createNodeBody struct {
	ParentID   ParentRef  `json:"parent_id,omitempty"`
	Name       string     `json:"name"`
	Note       string     `json:"note,omitempty"`
	LayoutMode LayoutMode `json:"layoutMode,omitempty"`
	Position   Position   `json:"position,omitempty"`
}

// CreateNodeResponse is the response from creating a node.
type CreateNodeResponse struct {
	ItemID string `json:"item_id"`
}

// NodesCreateCall creates a new node.
type NodesCreateCall struct {
	s    *service
	body *createNodeBody
}

// ParentID sets the parent for the new node.
func (c *NodesCreateCall) ParentID(p ParentRef) *NodesCreateCall {
	c.body.ParentID = p
	return c
}

// Note sets the note content.
func (c *NodesCreateCall) Note(note string) *NodesCreateCall {
	c.body.Note = note
	return c
}

// Layout sets the layout mode.
func (c *NodesCreateCall) Layout(mode LayoutMode) *NodesCreateCall {
	c.body.LayoutMode = mode
	return c
}

// Position sets the insertion position.
func (c *NodesCreateCall) Position(pos Position) *NodesCreateCall {
	c.body.Position = pos
	return c
}

// Do executes the create call.
func (c *NodesCreateCall) Do(ctx context.Context) (*CreateNodeResponse, error) {
	req, err := c.s.newRequest("POST", "/nodes", c.body)
	if err != nil {
		return nil, err
	}
	var resp CreateNodeResponse
	if err := c.s.do(ctx, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// --- Update ---

type updateNodeBody struct {
	Name       *string     `json:"name,omitempty"`
	Note       *string     `json:"note,omitempty"`
	LayoutMode *LayoutMode `json:"layoutMode,omitempty"`
}

// NodesUpdateCall updates an existing node.
type NodesUpdateCall struct {
	s    *service
	id   string
	body *updateNodeBody
}

// Name sets the new name.
func (c *NodesUpdateCall) Name(name string) *NodesUpdateCall {
	c.body.Name = &name
	return c
}

// Note sets the new note.
func (c *NodesUpdateCall) Note(note string) *NodesUpdateCall {
	c.body.Note = &note
	return c
}

// Layout sets the new layout mode.
func (c *NodesUpdateCall) Layout(mode LayoutMode) *NodesUpdateCall {
	c.body.LayoutMode = &mode
	return c
}

// Do executes the update call.
func (c *NodesUpdateCall) Do(ctx context.Context) error {
	req, err := c.s.newRequest("POST", "/nodes/"+c.id, c.body)
	if err != nil {
		return err
	}
	return c.s.do(ctx, req, nil)
}

// --- Delete ---

// NodesDeleteCall permanently deletes a node.
type NodesDeleteCall struct {
	s  *service
	id string
}

// Do executes the delete call.
func (c *NodesDeleteCall) Do(ctx context.Context) error {
	req, err := c.s.newRequest("DELETE", "/nodes/"+c.id, nil)
	if err != nil {
		return err
	}
	return c.s.do(ctx, req, nil)
}

// --- Move ---

type moveNodeBody struct {
	ParentID ParentRef `json:"parent_id,omitempty"`
	Position Position  `json:"position,omitempty"`
}

// NodesMoveCall moves a node to a new parent.
type NodesMoveCall struct {
	s    *service
	id   string
	body *moveNodeBody
}

// ParentID sets the new parent.
func (c *NodesMoveCall) ParentID(p ParentRef) *NodesMoveCall {
	c.body.ParentID = p
	return c
}

// Position sets the position within the new parent.
func (c *NodesMoveCall) Position(pos Position) *NodesMoveCall {
	c.body.Position = pos
	return c
}

// Do executes the move call.
func (c *NodesMoveCall) Do(ctx context.Context) error {
	req, err := c.s.newRequest("POST", "/nodes/"+c.id+"/move", c.body)
	if err != nil {
		return err
	}
	return c.s.do(ctx, req, nil)
}

// --- Complete ---

// NodesCompleteCall marks a node as completed.
type NodesCompleteCall struct {
	s  *service
	id string
}

// Do executes the complete call.
func (c *NodesCompleteCall) Do(ctx context.Context) error {
	req, err := c.s.newRequest("POST", "/nodes/"+c.id+"/complete", nil)
	if err != nil {
		return err
	}
	return c.s.do(ctx, req, nil)
}

// --- Uncomplete ---

// NodesUncompleteCall marks a node as not completed.
type NodesUncompleteCall struct {
	s  *service
	id string
}

// Do executes the uncomplete call.
func (c *NodesUncompleteCall) Do(ctx context.Context) error {
	req, err := c.s.newRequest("POST", "/nodes/"+c.id+"/uncomplete", nil)
	if err != nil {
		return err
	}
	return c.s.do(ctx, req, nil)
}

// --- Export ---

// NodesExportCall exports all nodes as a flat list.
// Note: This endpoint is rate-limited to 1 request per minute.
type NodesExportCall struct {
	s *service
}

// Do executes the export call.
func (c *NodesExportCall) Do(ctx context.Context) ([]*Node, error) {
	req, err := c.s.newRequest("GET", "/nodes-export", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Nodes []*Node `json:"nodes"`
	}
	if err := c.s.do(ctx, req, &resp); err != nil {
		return nil, err
	}
	return resp.Nodes, nil
}
