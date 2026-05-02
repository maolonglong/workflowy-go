package workflowy

import (
	"strings"
	"testing"
)

const (
	testFullID  = "3495d784-5db2-408f-8c4a-7ae1be810d4f"
	testShortID = "7ae1be810d4f"
)

func TestSanitizeNodeID(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "full uuid", in: testFullID, want: testFullID},
		{name: "internal link", in: "https://workflowy.com/#/" + testShortID, want: testShortID},
		{name: "fragment only", in: "#/" + testShortID, want: testShortID},
		{name: "surrounding whitespace", in: "  " + testFullID + "  ", want: testFullID},
		{name: "none", in: "None", want: "None"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := SanitizeNodeID(tc.in); got != tc.want {
				t.Fatalf("SanitizeNodeID(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestIDShapeChecks(t *testing.T) {
	if !IsShortID(testShortID) {
		t.Fatalf("IsShortID(%q) = false, want true", testShortID)
	}
	if IsShortID(testFullID) {
		t.Fatalf("IsShortID(%q) = true, want false", testFullID)
	}
	if !IsNodeUUID(testFullID) {
		t.Fatalf("IsNodeUUID(%q) = false, want true", testFullID)
	}
	if IsNodeUUID(testShortID) {
		t.Fatalf("IsNodeUUID(%q) = true, want false", testShortID)
	}
}

func TestResolveShortIDFromNodes(t *testing.T) {
	nodes := []*Node{
		{ID: "11111111-1111-1111-1111-111111111111"},
		{ID: testFullID},
	}

	got, err := ResolveShortIDFromNodes(nodes, strings.ToUpper(testShortID))
	if err != nil {
		t.Fatalf("ResolveShortIDFromNodes() error = %v", err)
	}
	if got != testFullID {
		t.Fatalf("ResolveShortIDFromNodes() = %q, want %q", got, testFullID)
	}
}

func TestResolveShortIDFromNodesRejectsAmbiguousMatches(t *testing.T) {
	nodes := []*Node{
		{ID: testFullID},
		{ID: "aaaaaaaa-aaaa-aaaa-aaaa-" + testShortID},
	}

	_, err := ResolveShortIDFromNodes(nodes, testShortID)
	if err == nil {
		t.Fatal("ResolveShortIDFromNodes() unexpectedly succeeded")
	}
	if !strings.Contains(err.Error(), "multiple nodes match short ID") {
		t.Fatalf("unexpected error: %v", err)
	}
}
