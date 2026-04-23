package workflowy

import (
	"encoding/json"
	"fmt"
	"time"
)

// Node represents a single bullet point in WorkFlowy.
type Node struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Note        *string    `json:"note"`
	Priority    float64    `json:"priority"`
	Data        NodeData   `json:"data"`
	CreatedAt   Timestamp  `json:"createdAt"`
	ModifiedAt  Timestamp  `json:"modifiedAt"`
	CompletedAt *Timestamp `json:"completedAt"`

	// Only present in Export responses.
	ParentID  *string `json:"parent_id,omitempty"`
	Completed *bool   `json:"completed,omitempty"`
}

// NodeData holds display metadata for a node.
type NodeData struct {
	LayoutMode LayoutMode `json:"layoutMode"`
}

// Target represents a quick-access entry point (e.g., inbox, home).
type Target struct {
	Key  string     `json:"key"`
	Type TargetType `json:"type"`
	Name *string    `json:"name"`
}

// LayoutMode defines the display mode of a node.
type LayoutMode string

const (
	LayoutBullets    LayoutMode = "bullets"
	LayoutTodo       LayoutMode = "todo"
	LayoutH1         LayoutMode = "h1"
	LayoutH2         LayoutMode = "h2"
	LayoutH3         LayoutMode = "h3"
	LayoutCodeBlock  LayoutMode = "code-block"
	LayoutQuoteBlock LayoutMode = "quote-block"
)

// TargetType defines the type of a Target.
type TargetType string

const (
	TargetTypeShortcut TargetType = "shortcut"
	TargetTypeSystem   TargetType = "system"
)

// Position defines where a new node is placed among siblings.
type Position string

const (
	PositionTop    Position = "top"
	PositionBottom Position = "bottom"
)

// ParentRef identifies the parent of a node.
// It can be a node UUID, a target key, or "None" for top-level.
type ParentRef string

const (
	ParentNone  ParentRef = "None"
	TargetHome  ParentRef = "home"
	TargetInbox ParentRef = "inbox"
)

// NodeParent creates a ParentRef from a node ID.
func NodeParent(id string) ParentRef {
	return ParentRef(id)
}

// Timestamp wraps time.Time with Unix timestamp JSON serialization.
type Timestamp struct {
	time.Time
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Unix())
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	var unix int64
	if string(data) == "null" {
		return nil
	}
	if err := json.Unmarshal(data, &unix); err != nil {
		return fmt.Errorf("workflowy: invalid timestamp: %w", err)
	}
	t.Time = time.Unix(unix, 0)
	return nil
}
