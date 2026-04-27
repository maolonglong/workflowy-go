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
		{
			name:  "matches needle only present in note",
			query: `overview`,
			want:  []string{"project-root"},
		},
		{
			name:  "case insensitive needle match",
			query: `PROJECT`,
			want:  []string{"project-root"},
		},
		{
			name:  "negated quoted phrase excludes match",
			query: `is:todo -"follow up soon"`,
			want:  []string{"done-task", "deep-task"},
		},
		{
			name:  "lowercase or operator behaves like OR",
			query: `@me or is:code-block`,
			want:  []string{"open-task", "code-snippet", "deep-task"},
		},
		{
			name:  "multi clause or preserves input order",
			query: `is:code-block OR @me OR @you`,
			want:  []string{"open-task", "done-task", "code-snippet", "deep-task"},
		},
		{
			name:  "double negation matches positively",
			query: `--@me`,
			want:  []string{"open-task", "deep-task"},
		},
		{
			name:  "created uses week unit",
			query: `created:1w`,
			want:  []string{"open-task", "code-snippet", "frontend", "deep-task"},
		},
		{
			name:  "changed uses hour unit",
			query: `changed:1h`,
			want:  []string{"deep-task"},
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
		{query: `"unterminated`, wantErr: `unterminated quoted phrase`},
		{query: `""`, wantErr: `empty quoted phrase`},
		{query: `   `, wantErr: `query is empty`},
		{query: `> foo`, wantErr: `empty segment before ">"`},
		{query: `-`, wantErr: `dangling "-"`},
		{query: `or project`, wantErr: `missing search term before "or"`},
		{query: `in:inbox`, wantErr: `in: is not supported`},
		{query: `highlight:yellow`, wantErr: `highlight: is not supported`},
		{query: `date-after:2024`, wantErr: `date-after: is not supported`},
		{query: `foo:bar`, wantErr: `unsupported operator "foo:"`},
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

func TestSearchQueryExecutionEdgeCases(t *testing.T) {
	fixedNow := time.Unix(1760000000, 0).UTC()
	restore := stubSearchNow(fixedNow)
	t.Cleanup(restore)

	t.Run("is:complete via completedAt when completed flag is nil", func(t *testing.T) {
		ts := workflowy.Timestamp{Time: fixedNow}
		node := &workflowy.Node{
			ID:          "x",
			Name:        "Done",
			Data:        workflowy.NodeData{LayoutMode: workflowy.LayoutBullets},
			CreatedAt:   ts,
			ModifiedAt:  ts,
			CompletedAt: &ts,
		}
		matches := mustCompile(t, `is:complete`).Filter([]*workflowy.Node{node})
		if got, want := nodeIDs(matches), []string{"x"}; !equalStrings(got, want) {
			t.Fatalf("got %v want %v", got, want)
		}
	})

	t.Run("is:complete via completed flag without completedAt", func(t *testing.T) {
		completed := true
		ts := workflowy.Timestamp{Time: fixedNow}
		node := &workflowy.Node{
			ID:         "x",
			Name:       "Done",
			Data:       workflowy.NodeData{LayoutMode: workflowy.LayoutBullets},
			CreatedAt:  ts,
			ModifiedAt: ts,
			Completed:  &completed,
		}
		matches := mustCompile(t, `is:complete`).Filter([]*workflowy.Node{node})
		if got, want := nodeIDs(matches), []string{"x"}; !equalStrings(got, want) {
			t.Fatalf("got %v want %v", got, want)
		}
	})

	t.Run("is:complete is false when completed flag is false and completedAt nil", func(t *testing.T) {
		completed := false
		ts := workflowy.Timestamp{Time: fixedNow}
		node := &workflowy.Node{
			ID:         "x",
			Name:       "Open",
			Data:       workflowy.NodeData{LayoutMode: workflowy.LayoutBullets},
			CreatedAt:  ts,
			ModifiedAt: ts,
			Completed:  &completed,
		}
		matches := mustCompile(t, `is:complete`).Filter([]*workflowy.Node{node})
		if len(matches) != 0 {
			t.Fatalf("expected no matches, got %v", nodeIDs(matches))
		}
	})

	t.Run("layout matchers cover all variants", func(t *testing.T) {
		layouts := map[string]workflowy.LayoutMode{
			"is:bullets":     workflowy.LayoutBullets,
			"is:todo":        workflowy.LayoutTodo,
			"is:h1":          workflowy.LayoutH1,
			"is:h2":          workflowy.LayoutH2,
			"is:h3":          workflowy.LayoutH3,
			"is:code-block":  workflowy.LayoutCodeBlock,
			"is:quote-block": workflowy.LayoutQuoteBlock,
		}
		ts := workflowy.Timestamp{Time: fixedNow}
		var nodes []*workflowy.Node
		for q, layout := range layouts {
			nodes = append(nodes, &workflowy.Node{
				ID:         q,
				Name:       q,
				Data:       workflowy.NodeData{LayoutMode: layout},
				CreatedAt:  ts,
				ModifiedAt: ts,
			})
		}
		for query := range layouts {
			matches := mustCompile(t, query).Filter(nodes)
			if got, want := nodeIDs(matches), []string{query}; !equalStrings(got, want) {
				t.Fatalf("query %q got %v want %v", query, got, want)
			}
		}
	})

	t.Run("recent timestamp ignores zero time", func(t *testing.T) {
		node := &workflowy.Node{
			ID:   "z",
			Name: "Zero",
			Data: workflowy.NodeData{LayoutMode: workflowy.LayoutBullets},
		}
		for _, q := range []string{`created:99w`, `changed:99w`} {
			matches := mustCompile(t, q).Filter([]*workflowy.Node{node})
			if len(matches) != 0 {
				t.Fatalf("query %q expected no matches, got %v", q, nodeIDs(matches))
			}
		}
	})

	t.Run("recent timestamp matches at boundary", func(t *testing.T) {
		// created:1h means since = now - 1h. Equal timestamp must match (uses !Before).
		boundary := workflowy.Timestamp{Time: fixedNow.Add(-time.Hour)}
		justBefore := workflowy.Timestamp{Time: fixedNow.Add(-time.Hour - time.Second)}
		nodes := []*workflowy.Node{
			{ID: "boundary", Name: "B", Data: workflowy.NodeData{LayoutMode: workflowy.LayoutBullets}, CreatedAt: boundary, ModifiedAt: boundary},
			{ID: "stale", Name: "S", Data: workflowy.NodeData{LayoutMode: workflowy.LayoutBullets}, CreatedAt: justBefore, ModifiedAt: justBefore},
		}
		matches := mustCompile(t, `created:1h`).Filter(nodes)
		if got, want := nodeIDs(matches), []string{"boundary"}; !equalStrings(got, want) {
			t.Fatalf("got %v want %v", got, want)
		}
	})

	t.Run("nested filter skips when ancestor is missing from lookup", func(t *testing.T) {
		missing := "missing"
		ts := workflowy.Timestamp{Time: fixedNow}
		node := &workflowy.Node{
			ID:         "child",
			Name:       "child @me",
			Data:       workflowy.NodeData{LayoutMode: workflowy.LayoutTodo},
			ParentID:   &missing,
			CreatedAt:  ts,
			ModifiedAt: ts,
		}
		matches := mustCompile(t, `#project > @me`).Filter([]*workflowy.Node{node})
		if len(matches) != 0 {
			t.Fatalf("expected no matches when ancestor lookup fails, got %v", nodeIDs(matches))
		}
	})

	t.Run("self cycle does not loop and does not match nested ancestor", func(t *testing.T) {
		self := "self"
		ts := workflowy.Timestamp{Time: fixedNow}
		node := &workflowy.Node{
			ID:         self,
			Name:       "Task @me",
			Data:       workflowy.NodeData{LayoutMode: workflowy.LayoutTodo},
			ParentID:   &self,
			CreatedAt:  ts,
			ModifiedAt: ts,
		}
		matches := mustCompile(t, `Task > @me`).Filter([]*workflowy.Node{node})
		if len(matches) != 0 {
			t.Fatalf("expected no matches, got %v", nodeIDs(matches))
		}
	})

	t.Run("filter on empty input returns empty result", func(t *testing.T) {
		matches := mustCompile(t, `anything`).Filter(nil)
		if len(matches) != 0 {
			t.Fatalf("expected empty result, got %v", nodeIDs(matches))
		}
	})

	t.Run("intermediate non-matching segment fails nested chain", func(t *testing.T) {
		ts := workflowy.Timestamp{Time: fixedNow}
		root := &workflowy.Node{ID: "root", Name: "#project", Data: workflowy.NodeData{LayoutMode: workflowy.LayoutBullets}, CreatedAt: ts, ModifiedAt: ts}
		rootID := root.ID
		mid := &workflowy.Node{ID: "mid", Name: "Other", Data: workflowy.NodeData{LayoutMode: workflowy.LayoutBullets}, CreatedAt: ts, ModifiedAt: ts, ParentID: &rootID}
		midID := mid.ID
		leaf := &workflowy.Node{ID: "leaf", Name: "@me", Data: workflowy.NodeData{LayoutMode: workflowy.LayoutTodo}, CreatedAt: ts, ModifiedAt: ts, ParentID: &midID}

		matches := mustCompile(t, `#project > frontend > @me`).Filter([]*workflowy.Node{root, mid, leaf})
		if len(matches) != 0 {
			t.Fatalf("expected no matches when intermediate segment is absent, got %v", nodeIDs(matches))
		}
	})
}

func FuzzCompileSearchQuery(f *testing.F) {
	seeds := []string{
		`#project > is:todo -is:complete`,
		`"follow up soon"`,
		`@me OR is:code-block`,
		`#project > frontend > @me`,
		`has:note`,
		`is:complete`,
		`created:7d`,
		`changed:1h`,
		`--@me`,
		`> foo`,
		`project >`,
		`"unterminated`,
		`""`,
		``,
		`   `,
		`-`,
		`foo:bar`,
		`OR project`,
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		q, err := compileSearchQuery(raw)
		if err != nil {
			return
		}
		// A successfully compiled query must be safe to execute against
		// any node set, including nil and self-referential cycles.
		_ = q.Filter(nil)
		self := "self"
		ts := workflowy.Timestamp{Time: time.Unix(1760000000, 0).UTC()}
		_ = q.Filter([]*workflowy.Node{{
			ID:         self,
			Name:       raw,
			Data:       workflowy.NodeData{LayoutMode: workflowy.LayoutBullets},
			ParentID:   &self,
			CreatedAt:  ts,
			ModifiedAt: ts,
		}})
	})
}

func mustCompile(t *testing.T, raw string) *compiledSearchQuery {
	t.Helper()
	q, err := compileSearchQuery(raw)
	if err != nil {
		t.Fatalf("compileSearchQuery(%q) returned error: %v", raw, err)
	}
	return q
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
