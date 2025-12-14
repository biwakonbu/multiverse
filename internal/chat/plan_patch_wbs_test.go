package chat

import (
	"testing"

	"github.com/biwakonbu/agent-runner/internal/meta"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PRD 13.3 #4: WBS 不変条件テスト (delete/move/cascade の回帰防止)

// TestRemoveNodeFromWBS_CascadeFalse_ReparentsChildren tests that cascade=false
// moves children to deleted node's parent (案A: splice 実装)
func TestRemoveNodeFromWBS_CascadeFalse_ReparentsChildren(t *testing.T) {
	// Setup: P -> X -> [a, b, c]
	parentID := "parent-P"
	rootID := "node-root"
	wbs := &persistence.WBS{
		RootNodeID: rootID,
		NodeIndex: []persistence.NodeIndex{
			{NodeID: rootID, ParentID: nil, Children: []string{parentID}},
			{NodeID: parentID, ParentID: &rootID, Children: []string{"X"}},
			{NodeID: "X", ParentID: &parentID, Children: []string{"a", "b", "c"}},
			{NodeID: "a", ParentID: strPtr("X"), Children: []string{}},
			{NodeID: "b", ParentID: strPtr("X"), Children: []string{}},
			{NodeID: "c", ParentID: strPtr("X"), Children: []string{}},
		},
	}

	// Act: Delete X with cascade=false
	removeNodeFromWBS(wbs, "X", false)

	// Assert: X is removed
	for _, n := range wbs.NodeIndex {
		assert.NotEqual(t, "X", n.NodeID, "X should be removed from WBS")
	}

	// Assert: P's children now includes a, b, c at X's position
	pNode := findNode(wbs, parentID)
	require.NotNil(t, pNode, "parent P should still exist")
	assert.Equal(t, []string{"a", "b", "c"}, pNode.Children, "children should be spliced to parent")

	// Assert: a, b, c now have parent P
	for _, childID := range []string{"a", "b", "c"} {
		child := findNode(wbs, childID)
		require.NotNil(t, child, "%s should still exist", childID)
		assert.Equal(t, parentID, *child.ParentID, "%s should have parent P", childID)
	}
}

// TestRemoveNodeFromWBS_CascadeTrue_RemovesFromParent tests that cascade=true
// simply removes the node from its parent's children list (subtree is removed by collectDeleteIDs)
func TestRemoveNodeFromWBS_CascadeTrue_RemovesFromParent(t *testing.T) {
	// Setup: P -> X -> [a, b]
	parentID := "parent-P"
	rootID := "node-root"
	wbs := &persistence.WBS{
		RootNodeID: rootID,
		NodeIndex: []persistence.NodeIndex{
			{NodeID: rootID, ParentID: nil, Children: []string{parentID}},
			{NodeID: parentID, ParentID: &rootID, Children: []string{"X", "Y"}},
			{NodeID: "X", ParentID: &parentID, Children: []string{"a", "b"}},
			{NodeID: "Y", ParentID: &parentID, Children: []string{}},
			{NodeID: "a", ParentID: strPtr("X"), Children: []string{}},
			{NodeID: "b", ParentID: strPtr("X"), Children: []string{}},
		},
	}

	// Act: Delete X with cascade=true (just removes from parent, subtree removal handled separately)
	removeNodeFromWBS(wbs, "X", true)

	// Assert: X is removed
	xNode := findNode(wbs, "X")
	assert.Nil(t, xNode, "X should be removed from WBS")

	// Assert: P's children no longer includes X (children a,b not spliced)
	pNode := findNode(wbs, parentID)
	require.NotNil(t, pNode, "parent P should still exist")
	assert.NotContains(t, pNode.Children, "X", "X should be removed from parent's children")
	assert.Contains(t, pNode.Children, "Y", "Y should still be in parent's children")
}

// TestWBSInvariants_NoOrphanNodes tests that all nodes (except root) have a valid parent
func TestWBSInvariants_NoOrphanNodes(t *testing.T) {
	rootID := "node-root"
	wbs := &persistence.WBS{
		RootNodeID: rootID,
		NodeIndex: []persistence.NodeIndex{
			{NodeID: rootID, ParentID: nil, Children: []string{"A", "B"}},
			{NodeID: "A", ParentID: &rootID, Children: []string{"A1"}},
			{NodeID: "A1", ParentID: strPtr("A"), Children: []string{}},
			{NodeID: "B", ParentID: &rootID, Children: []string{}},
		},
	}

	// Delete A with cascade=false -> A1 should be reparented to root
	removeNodeFromWBS(wbs, "A", false)

	err := validateWBSInvariants(wbs)
	assert.NoError(t, err, "WBS should have no orphan nodes after delete cascade=false")
}

// TestWBSInvariants_NoDuplicateChildren tests that no node has duplicate children
func TestWBSInvariants_NoDuplicateChildren(t *testing.T) {
	rootID := "node-root"
	wbs := &persistence.WBS{
		RootNodeID: rootID,
		NodeIndex: []persistence.NodeIndex{
			{NodeID: rootID, ParentID: nil, Children: []string{"A", "B"}},
			{NodeID: "A", ParentID: &rootID, Children: []string{}},
			{NodeID: "B", ParentID: &rootID, Children: []string{}},
		},
	}

	err := validateWBSInvariants(wbs)
	assert.NoError(t, err, "WBS should have no duplicate children")
}

// Helper functions

func findNode(wbs *persistence.WBS, nodeID string) *persistence.NodeIndex {
	for i := range wbs.NodeIndex {
		if wbs.NodeIndex[i].NodeID == nodeID {
			return &wbs.NodeIndex[i]
		}
	}
	return nil
}

func strPtr(s string) *string {
	return &s
}

// validateWBSInvariants checks PRD 11.2 invariants
func validateWBSInvariants(wbs *persistence.WBS) error {
	if wbs == nil {
		return nil
	}

	nodeByID := make(map[string]*persistence.NodeIndex)
	for i := range wbs.NodeIndex {
		nodeByID[wbs.NodeIndex[i].NodeID] = &wbs.NodeIndex[i]
	}

	// Invariant 1: root_node_id must exist in node_index
	if _, ok := nodeByID[wbs.RootNodeID]; !ok {
		return &invariantError{"root_node_id not in node_index", wbs.RootNodeID}
	}

	// Invariant 2: All nodes (except root) have parent_id in children
	for _, n := range wbs.NodeIndex {
		if n.NodeID == wbs.RootNodeID {
			continue
		}
		if n.ParentID == nil || *n.ParentID == "" {
			return &invariantError{"non-root node has no parent", n.NodeID}
		}
		parent, ok := nodeByID[*n.ParentID]
		if !ok {
			return &invariantError{"parent not found", *n.ParentID}
		}
		found := false
		for _, c := range parent.Children {
			if c == n.NodeID {
				found = true
				break
			}
		}
		if !found {
			return &invariantError{"node not in parent's children", n.NodeID}
		}
	}

	// Invariant 3: No duplicate children
	for _, n := range wbs.NodeIndex {
		seen := make(map[string]struct{})
		for _, c := range n.Children {
			if _, ok := seen[c]; ok {
				return &invariantError{"duplicate child", c}
			}
			seen[c] = struct{}{}
		}
	}

	return nil
}

type invariantError struct {
	msg    string
	nodeID string
}

func (e *invariantError) Error() string {
	return e.msg + ": " + e.nodeID
}

// QH-003 (PRD 13.3 #2): move 回帰防止テスト

// TestMoveNodeInWBS_BasicMove tests moving a node to a new parent
func TestMoveNodeInWBS_BasicMove(t *testing.T) {
	// Setup: root -> [A -> [A1], B]
	rootID := "node-root"
	wbs := &persistence.WBS{
		RootNodeID: rootID,
		NodeIndex: []persistence.NodeIndex{
			{NodeID: rootID, ParentID: nil, Children: []string{"A", "B"}},
			{NodeID: "A", ParentID: &rootID, Children: []string{"A1"}},
			{NodeID: "A1", ParentID: strPtr("A"), Children: []string{}},
			{NodeID: "B", ParentID: &rootID, Children: []string{}},
		},
	}

	// Move A1 to B
	op := meta.PlanOperation{ParentID: strPtr("B")}
	err := moveNodeInWBS(wbs, "A1", op, nil)
	require.NoError(t, err)

	// Validate invariants
	err = validateWBSInvariants(wbs)
	require.NoError(t, err, "WBS should maintain invariants after move")

	// Assert: A1's new parent is B
	a1Node := findNode(wbs, "A1")
	require.NotNil(t, a1Node)
	assert.Equal(t, "B", *a1Node.ParentID, "A1 should have B as parent")

	// Assert: A1 is in B's children
	bNode := findNode(wbs, "B")
	require.NotNil(t, bNode)
	assert.Contains(t, bNode.Children, "A1", "B should have A1 in children")

	// Assert: A1 is removed from A's children
	aNode := findNode(wbs, "A")
	require.NotNil(t, aNode)
	assert.NotContains(t, aNode.Children, "A1", "A should not have A1 in children anymore")
}

// TestMoveNodeInWBS_WithPosition tests moving with position (before/after/index)
func TestMoveNodeInWBS_WithPosition(t *testing.T) {
	rootID := "node-root"
	wbs := &persistence.WBS{
		RootNodeID: rootID,
		NodeIndex: []persistence.NodeIndex{
			{NodeID: rootID, ParentID: nil, Children: []string{"A", "B", "C"}},
			{NodeID: "A", ParentID: &rootID, Children: []string{}},
			{NodeID: "B", ParentID: &rootID, Children: []string{}},
			{NodeID: "C", ParentID: &rootID, Children: []string{}},
			{NodeID: "X", ParentID: &rootID, Children: []string{}},
		},
	}

	// Move X to root at position before B
	op := meta.PlanOperation{
		ParentID: &rootID,
		Position: &meta.WBSPosition{Before: "B"},
	}
	err := moveNodeInWBS(wbs, "X", op, nil)
	require.NoError(t, err)

	err = validateWBSInvariants(wbs)
	require.NoError(t, err)

	// Assert: root's children order should be [A, X, B, C]
	rootNode := findNode(wbs, rootID)
	require.NotNil(t, rootNode)
	assert.Equal(t, []string{"A", "X", "B", "C"}, rootNode.Children, "X should be inserted before B")
}

// TestMoveNodeInWBS_WithPositionAfter tests moving with position after
func TestMoveNodeInWBS_WithPositionAfter(t *testing.T) {
	rootID := "node-root"
	wbs := &persistence.WBS{
		RootNodeID: rootID,
		NodeIndex: []persistence.NodeIndex{
			{NodeID: rootID, ParentID: nil, Children: []string{"A", "B", "C"}},
			{NodeID: "A", ParentID: &rootID, Children: []string{}},
			{NodeID: "B", ParentID: &rootID, Children: []string{}},
			{NodeID: "C", ParentID: &rootID, Children: []string{}},
			{NodeID: "X", ParentID: &rootID, Children: []string{}},
		},
	}

	// Move X to root at position after A
	op := meta.PlanOperation{
		ParentID: &rootID,
		Position: &meta.WBSPosition{After: "A"},
	}
	err := moveNodeInWBS(wbs, "X", op, nil)
	require.NoError(t, err)

	err = validateWBSInvariants(wbs)
	require.NoError(t, err)

	rootNode := findNode(wbs, rootID)
	require.NotNil(t, rootNode)
	assert.Equal(t, []string{"A", "X", "B", "C"}, rootNode.Children, "X should be inserted after A")
}

// TestMoveNodeInWBS_WithPositionIndex tests moving with position index
func TestMoveNodeInWBS_WithPositionIndex(t *testing.T) {
	rootID := "node-root"
	wbs := &persistence.WBS{
		RootNodeID: rootID,
		NodeIndex: []persistence.NodeIndex{
			{NodeID: rootID, ParentID: nil, Children: []string{"A", "B", "C"}},
			{NodeID: "A", ParentID: &rootID, Children: []string{}},
			{NodeID: "B", ParentID: &rootID, Children: []string{}},
			{NodeID: "C", ParentID: &rootID, Children: []string{}},
			{NodeID: "X", ParentID: &rootID, Children: []string{}},
		},
	}

	// Move X to root at index 0
	idx := 0
	op := meta.PlanOperation{
		ParentID: &rootID,
		Position: &meta.WBSPosition{Index: &idx},
	}
	err := moveNodeInWBS(wbs, "X", op, nil)
	require.NoError(t, err)

	err = validateWBSInvariants(wbs)
	require.NoError(t, err)

	rootNode := findNode(wbs, rootID)
	require.NotNil(t, rootNode)
	assert.Equal(t, []string{"X", "A", "B", "C"}, rootNode.Children, "X should be at index 0")
}

// TestMoveNodeInWBS_SameParent tests moving within the same parent (reorder)
func TestMoveNodeInWBS_SameParent(t *testing.T) {
	rootID := "node-root"
	wbs := &persistence.WBS{
		RootNodeID: rootID,
		NodeIndex: []persistence.NodeIndex{
			{NodeID: rootID, ParentID: nil, Children: []string{"A", "B", "C"}},
			{NodeID: "A", ParentID: &rootID, Children: []string{}},
			{NodeID: "B", ParentID: &rootID, Children: []string{}},
			{NodeID: "C", ParentID: &rootID, Children: []string{}},
		},
	}

	// Move C to before A (still under root)
	op := meta.PlanOperation{
		ParentID: &rootID,
		Position: &meta.WBSPosition{Before: "A"},
	}
	err := moveNodeInWBS(wbs, "C", op, nil)
	require.NoError(t, err)

	err = validateWBSInvariants(wbs)
	require.NoError(t, err)

	rootNode := findNode(wbs, rootID)
	require.NotNil(t, rootNode)
	assert.Equal(t, []string{"C", "A", "B"}, rootNode.Children, "C should be moved to before A")
}

// TestMoveNodeInWBS_NoDuplicateChildren ensures no duplicate children after multiple moves
func TestMoveNodeInWBS_NoDuplicateChildren(t *testing.T) {
	rootID := "node-root"
	wbs := &persistence.WBS{
		RootNodeID: rootID,
		NodeIndex: []persistence.NodeIndex{
			{NodeID: rootID, ParentID: nil, Children: []string{"A", "B"}},
			{NodeID: "A", ParentID: &rootID, Children: []string{"X"}},
			{NodeID: "B", ParentID: &rootID, Children: []string{}},
			{NodeID: "X", ParentID: strPtr("A"), Children: []string{}},
		},
	}

	// Move X to B, then back to A
	op1 := meta.PlanOperation{ParentID: strPtr("B")}
	err := moveNodeInWBS(wbs, "X", op1, nil)
	require.NoError(t, err)

	op2 := meta.PlanOperation{ParentID: strPtr("A")}
	err = moveNodeInWBS(wbs, "X", op2, nil)
	require.NoError(t, err)

	err = validateWBSInvariants(wbs)
	require.NoError(t, err, "no orphans or duplicates after multiple moves")
}
