package workflowy

import "sort"

// NodeTree is a node with its children, used for tree reconstruction.
type NodeTree struct {
	*Node
	Children []*NodeTree
}

// BuildTree reconstructs a tree from a flat list of exported nodes.
// Nodes are sorted by priority within each level.
func BuildTree(nodes []*Node) []*NodeTree {
	lookup := make(map[string]*NodeTree, len(nodes))
	for _, n := range nodes {
		lookup[n.ID] = &NodeTree{Node: n}
	}

	var roots []*NodeTree
	for _, n := range nodes {
		tree := lookup[n.ID]
		if n.ParentID == nil || *n.ParentID == "" {
			roots = append(roots, tree)
		} else if parent, ok := lookup[*n.ParentID]; ok {
			parent.Children = append(parent.Children, tree)
		} else {
			roots = append(roots, tree)
		}
	}

	sortChildren(roots)
	return roots
}

func sortChildren(trees []*NodeTree) {
	sort.Slice(trees, func(i, j int) bool {
		return trees[i].Priority < trees[j].Priority
	})
	for _, t := range trees {
		if len(t.Children) > 0 {
			sortChildren(t.Children)
		}
	}
}
