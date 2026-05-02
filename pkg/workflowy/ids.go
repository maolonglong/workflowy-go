package workflowy

import (
	"fmt"
	"strings"
)

// SanitizeNodeID strips a WorkFlowy internal-link prefix and removes characters
// that cannot appear in a node UUID.
func SanitizeNodeID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" || id == string(ParentNone) {
		return id
	}
	if _, after, ok := strings.Cut(id, "#/"); ok {
		id = after
	}

	var b strings.Builder
	for _, r := range id {
		if isHexRune(r) || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// IsShortID reports whether id is a WorkFlowy internal-link short ID.
func IsShortID(id string) bool {
	if len(id) != 12 {
		return false
	}
	for _, r := range id {
		if !isHexRune(r) {
			return false
		}
	}
	return true
}

// IsNodeUUID reports whether id looks like a full WorkFlowy node UUID.
func IsNodeUUID(id string) bool {
	parts := strings.Split(id, "-")
	if len(parts) != 5 {
		return false
	}
	want := []int{8, 4, 4, 4, 12}
	for i, part := range parts {
		if len(part) != want[i] {
			return false
		}
		for _, r := range part {
			if !isHexRune(r) {
				return false
			}
		}
	}
	return true
}

// ResolveShortIDFromNodes resolves a 12-character short ID against exported nodes.
func ResolveShortIDFromNodes(nodes []*Node, shortID string) (string, error) {
	if !IsShortID(shortID) {
		return "", fmt.Errorf("workflowy: %q is not a short ID", shortID)
	}

	needle := strings.ToLower(shortID)
	var matches []string
	for _, node := range nodes {
		if node == nil {
			continue
		}
		if strings.HasSuffix(strings.ToLower(node.ID), needle) {
			matches = append(matches, node.ID)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("workflowy: no node found with short ID %s", shortID)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("workflowy: multiple nodes match short ID %s: %s", shortID, strings.Join(matches, ", "))
	}
}

func isHexRune(r rune) bool {
	return (r >= '0' && r <= '9') ||
		(r >= 'a' && r <= 'f') ||
		(r >= 'A' && r <= 'F')
}
