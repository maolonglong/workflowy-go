package cli

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/maolonglong/workflowy-go/pkg/workflowy"
)

func TestCompileSearchQueryAgainstMockExportData(t *testing.T) {
	fixedNow := time.Unix(1760000000, 0).UTC()
	restore := stubSearchNow(fixedNow)
	t.Cleanup(restore)

	nodes := mockExportNodes(t)

	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{
			name:  "implicit and nested filters",
			query: `#project > is:todo -is:complete`,
			want:  []string{"open-task", "deep-task"},
		},
		{
			name:  "quoted phrase",
			query: `"follow up soon"`,
			want:  []string{"open-task"},
		},
		{
			name:  "or expression",
			query: `@me OR is:code-block`,
			want:  []string{"open-task", "code-snippet", "deep-task"},
		},
		{
			name:  "nested any depth",
			query: `#project > frontend > @me`,
			want:  []string{"deep-task"},
		},
		{
			name:  "note presence includes empty note field",
			query: `has:note`,
			want:  []string{"project-root", "open-task", "code-snippet"},
		},
		{
			name:  "completed uses export completion state",
			query: `is:complete`,
			want:  []string{"done-task"},
		},
		{
			name:  "created age filter",
			query: `created:7d`,
			want:  []string{"open-task", "code-snippet", "frontend", "deep-task"},
		},
		{
			name:  "changed age filter",
			query: `changed:1d`,
			want:  []string{"open-task", "code-snippet", "deep-task"},
		},
		{
			name:  "combined nested and time filter",
			query: `#project > changed:1d is:todo`,
			want:  []string{"open-task", "deep-task"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			query, err := compileSearchQuery(tc.query)
			if err != nil {
				t.Fatalf("compileSearchQuery(%q) returned error: %v", tc.query, err)
			}

			matches := query.Filter(nodes)
			if got := nodeIDs(matches); !equalStrings(got, tc.want) {
				t.Fatalf("unexpected matches for %q: got %v want %v", tc.query, got, tc.want)
			}
		})
	}
}

func TestCompileSearchQueryRejectsUnsupportedOrInvalidInput(t *testing.T) {
	tests := []struct {
		query   string
		wantErr string
	}{
		{query: `has:file`, wantErr: `has:file is not supported`},
		{query: `text:red`, wantErr: `text: is not supported`},
		{query: `project >`, wantErr: `empty segment after ">"`},
		{query: `OR project`, wantErr: `missing search term before "OR"`},
		{query: `changed:yesterday`, wantErr: `changed:yesterday must use a relative age`},
		{query: `created:`, wantErr: `created: must use a relative age`},
	}

	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			_, err := compileSearchQuery(tc.query)
			if err == nil {
				t.Fatalf("compileSearchQuery(%q) unexpectedly succeeded", tc.query)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("unexpected error for %q: got %q want substring %q", tc.query, err.Error(), tc.wantErr)
			}
		})
	}
}

func TestCompileSearchQueryHandlesAncestorCycles(t *testing.T) {
	nodes := []*workflowy.Node{
		testNode("a", "#project", "overview", workflowy.LayoutBullets, "b", false, 1760000000, 1760000000, nil),
		testNode("b", "Loop", "", workflowy.LayoutBullets, "a", false, 1760000000, 1760000000, nil),
		testNode("c", "Task @me", "", workflowy.LayoutTodo, "a", false, 1760000000, 1760000000, nil),
	}

	query, err := compileSearchQuery(`#project > @me`)
	if err != nil {
		t.Fatalf("compileSearchQuery returned error: %v", err)
	}

	matches := query.Filter(nodes)
	if got, want := nodeIDs(matches), []string{"c"}; !equalStrings(got, want) {
		t.Fatalf("unexpected matches: got %v want %v", got, want)
	}
}

func mockExportNodes(t *testing.T) []*workflowy.Node {
	t.Helper()

	const exportJSON = `{
  "nodes": [
    {
      "id": "project-root",
      "name": "#project Alpha",
      "note": "Project overview",
      "parent_id": null,
      "priority": 100,
      "completed": false,
      "data": { "layoutMode": "bullets" },
      "createdAt": 1758000000,
      "modifiedAt": 1759500000,
      "completedAt": null
    },
    {
      "id": "open-task",
      "name": "Write draft @me",
      "note": "Follow up soon",
      "parent_id": "project-root",
      "priority": 200,
      "completed": false,
      "data": { "layoutMode": "todo" },
      "createdAt": 1759800000,
      "modifiedAt": 1759950000,
      "completedAt": null
    },
    {
      "id": "done-task",
      "name": "Ship draft @you",
      "note": null,
      "parent_id": "project-root",
      "priority": 300,
      "completed": true,
      "data": { "layoutMode": "todo" },
      "createdAt": 1759000000,
      "modifiedAt": 1759700000,
      "completedAt": 1759700000
    },
    {
      "id": "code-snippet",
      "name": "Code snippet",
      "note": "",
      "parent_id": "project-root",
      "priority": 400,
      "completed": false,
      "data": { "layoutMode": "code-block" },
      "createdAt": 1759990000,
      "modifiedAt": 1759992800,
      "completedAt": null
    },
    {
      "id": "frontend",
      "name": "Frontend",
      "note": null,
      "parent_id": "project-root",
      "priority": 500,
      "completed": false,
      "data": { "layoutMode": "bullets" },
      "createdAt": 1759900000,
      "modifiedAt": 1759903600,
      "completedAt": null
    },
    {
      "id": "deep-task",
      "name": "Review UI @me",
      "note": null,
      "parent_id": "frontend",
      "priority": 600,
      "completed": false,
      "data": { "layoutMode": "todo" },
      "createdAt": 1759950000,
      "modifiedAt": 1759996400,
      "completedAt": null
    }
  ]
}`

	var payload struct {
		Nodes []*workflowy.Node `json:"nodes"`
	}
	if err := json.Unmarshal([]byte(exportJSON), &payload); err != nil {
		t.Fatalf("failed to unmarshal mock export JSON: %v", err)
	}
	return payload.Nodes
}

func testNode(id, name, note string, layout workflowy.LayoutMode, parent string, completed bool, createdAt, modifiedAt int64, completedAt *int64) *workflowy.Node {
	node := &workflowy.Node{
		ID:         id,
		Name:       name,
		Data:       workflowy.NodeData{LayoutMode: layout},
		CreatedAt:  workflowy.Timestamp{Time: time.Unix(createdAt, 0).UTC()},
		ModifiedAt: workflowy.Timestamp{Time: time.Unix(modifiedAt, 0).UTC()},
		Completed:  &completed,
	}
	if parent != "" {
		node.ParentID = &parent
	}
	if note != "" {
		node.Note = &note
	}
	if completedAt != nil {
		ts := workflowy.Timestamp{Time: time.Unix(*completedAt, 0).UTC()}
		node.CompletedAt = &ts
	}
	return node
}

func stubSearchNow(now time.Time) func() {
	original := searchNow
	searchNow = func() time.Time { return now }
	return func() { searchNow = original }
}

func nodeIDs(nodes []*workflowy.Node) []string {
	ids := make([]string, 0, len(nodes))
	for _, node := range nodes {
		ids = append(ids, node.ID)
	}
	return ids
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
