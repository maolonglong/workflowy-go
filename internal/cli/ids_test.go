package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/maolonglong/workflowy-go/pkg/workflowy"
)

const (
	cliFullID  = "3495d784-5db2-408f-8c4a-7ae1be810d4f"
	cliShortID = "7ae1be810d4f"
	cliInboxID = "22222222-2222-2222-2222-222222222222"
)

func TestResolveNodeID(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	a := newResolverTestApp(t, []*workflowy.Node{{ID: cliFullID}})

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "full uuid", in: cliFullID, want: cliFullID},
		{name: "short id", in: cliShortID, want: cliFullID},
		{name: "internal link", in: "https://workflowy.com/#/" + cliShortID, want: cliFullID},
		{name: "target key", in: "inbox", want: "inbox"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := a.resolveNodeID(context.Background(), tc.in)
			if err != nil {
				t.Fatalf("resolveNodeID(%q) error = %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("resolveNodeID(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestResolveNodeUUIDResolvesTargetKey(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	a := newResolverTestApp(t, []*workflowy.Node{{ID: cliFullID}})

	got, err := a.resolveNodeUUID(context.Background(), "inbox")
	if err != nil {
		t.Fatalf("resolveNodeUUID() error = %v", err)
	}
	if got != cliInboxID {
		t.Fatalf("resolveNodeUUID() = %q, want %q", got, cliInboxID)
	}
}

func TestResolveNodeIDRejectsAmbiguousShortID(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	a := newResolverTestApp(t, []*workflowy.Node{
		{ID: cliFullID},
		{ID: "aaaaaaaa-aaaa-aaaa-aaaa-" + cliShortID},
	})

	_, err := a.resolveNodeID(context.Background(), cliShortID)
	if err == nil {
		t.Fatal("resolveNodeID() unexpectedly succeeded")
	}
	if !strings.Contains(err.Error(), "multiple nodes match short ID") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func newResolverTestApp(t *testing.T, nodes []*workflowy.Node) *app {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/targets", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{
			"targets": []map[string]any{
				{"key": "inbox", "type": "system", "name": "Inbox"},
			},
		})
	})
	mux.HandleFunc("/nodes-export", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{"nodes": nodes})
	})
	mux.HandleFunc("/nodes/inbox", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, map[string]any{
			"node": map[string]any{
				"id":         cliInboxID,
				"name":       "Inbox",
				"priority":   0,
				"data":       map[string]string{"layoutMode": "bullets"},
				"createdAt":  0,
				"modifiedAt": 0,
			},
		})
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := workflowy.NewClient(
		workflowy.WithAPIKey("test-api-key"),
		workflowy.WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return &app{client: client}
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("encode JSON: %v", err)
	}
}
