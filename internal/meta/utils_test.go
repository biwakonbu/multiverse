package meta

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTrimWBSNodesBFS_UnderLimit tests that nodes under the limit are returned unchanged
func TestTrimWBSNodesBFS_UnderLimit(t *testing.T) {
	nodes := []WBSNodeIndex{
		{NodeID: "root", ParentID: nil, Children: []string{"A"}},
		{NodeID: "A", ParentID: strPtr("root"), Children: []string{}},
	}

	result := trimWBSNodesBFS(nodes, "root", 200)

	assert.Equal(t, nodes, result, "nodes under limit should be returned unchanged")
}

// TestTrimWBSNodesBFS_OverLimit_BFSOrder tests deterministic BFS ordering when over limit
func TestTrimWBSNodesBFS_OverLimit_BFSOrder(t *testing.T) {
	// Create a tree: root -> A,B -> A1,A2,B1,B2
	root := "node-root"
	nodes := []WBSNodeIndex{
		{NodeID: root, ParentID: nil, Children: []string{"A", "B"}},
		{NodeID: "A", ParentID: &root, Children: []string{"A1", "A2"}},
		{NodeID: "B", ParentID: &root, Children: []string{"B1", "B2"}},
		{NodeID: "A1", ParentID: strPtr("A"), Children: []string{}},
		{NodeID: "A2", ParentID: strPtr("A"), Children: []string{}},
		{NodeID: "B1", ParentID: strPtr("B"), Children: []string{}},
		{NodeID: "B2", ParentID: strPtr("B"), Children: []string{}},
	}

	// Limit to 4 nodes
	result := trimWBSNodesBFS(nodes, root, 4)

	require.Len(t, result, 4, "should return exactly maxNodes")

	// BFS order: root, A, B, A1 (breadth-first)
	expectedIDs := []string{root, "A", "B", "A1"}
	actualIDs := make([]string, len(result))
	for i, n := range result {
		actualIDs[i] = n.NodeID
	}
	assert.Equal(t, expectedIDs, actualIDs, "BFS order should be deterministic")
}

// TestTrimWBSNodesBFS_EmptyNodes tests handling of empty input
func TestTrimWBSNodesBFS_EmptyNodes(t *testing.T) {
	result := trimWBSNodesBFS(nil, "root", 200)
	assert.Empty(t, result, "nil input should return empty slice")

	result2 := trimWBSNodesBFS([]WBSNodeIndex{}, "root", 200)
	assert.Empty(t, result2, "empty input should return empty slice")
}

// TestTrimWBSNodesBFS_RootNotFound tests that when root is missing from a large tree,
// BFS starts from nowhere and returns empty (edge case)
func TestTrimWBSNodesBFS_RootNotFound(t *testing.T) {
	// Need more than maxNodes to trigger BFS logic
	nodes := make([]WBSNodeIndex, 0, 250)
	for i := 0; i < 250; i++ {
		nodeID := "node-" + string(rune('A'+i%26)) + string(rune('0'+i/26))
		nodes = append(nodes, WBSNodeIndex{NodeID: nodeID, ParentID: strPtr("parent"), Children: []string{}})
	}

	// BFS starts from "missing-root" which doesn't exist
	result := trimWBSNodesBFS(nodes, "missing-root", 200)
	assert.Empty(t, result, "BFS from missing root should return empty slice")
}

// TestBuildPlanPatchUserPrompt_WBSTrimming tests that WBS is trimmed in the prompt
func TestBuildPlanPatchUserPrompt_WBSTrimming(t *testing.T) {
	// Create 250 nodes (over 200 limit)
	root := "node-root"
	nodes := []WBSNodeIndex{{NodeID: root, ParentID: nil, Children: make([]string, 0, 249)}}
	for i := 1; i < 250; i++ {
		nodeID := "node-" + string(rune('A'+i%26)) + string(rune('0'+i/26))
		nodes[0].Children = append(nodes[0].Children, nodeID)
		nodes = append(nodes, WBSNodeIndex{NodeID: nodeID, ParentID: &root, Children: []string{}})
	}

	req := &PlanPatchRequest{
		UserInput: "test",
		Context: PlanPatchContext{
			ExistingWBS: &WBSOverview{
				RootNodeID: root,
				NodeIndex:  nodes,
			},
		},
	}

	prompt := buildPlanPatchUserPrompt(req)

	// Should indicate trimming
	assert.Contains(t, prompt, "showing 200 of 250 nodes", "prompt should indicate trimming")
}

func strPtr(s string) *string {
	return &s
}

// TestBuildPlanPatchUserPrompt_FullContent tests that required fields are included
func TestBuildPlanPatchUserPrompt_FullContent(t *testing.T) {
	root := "node-root"
	req := &PlanPatchRequest{
		UserInput: "create a new feature",
		Context: PlanPatchContext{
			ExistingTasks: []ExistingTaskSummary{
				{ID: "task-1", Title: "Task One", Status: "RUNNING", PhaseName: "Design", Milestone: "M1"},
			},
			ExistingWBS: &WBSOverview{
				RootNodeID: root,
				NodeIndex: []WBSNodeIndex{
					{NodeID: root, ParentID: nil, Children: []string{"task-1"}},
					{NodeID: "task-1", ParentID: &root, Children: []string{}},
				},
			},
			ConversationHistory: []ConversationMessage{
				{Role: "user", Content: "hello"},
				{Role: "assistant", Content: "hi"},
			},
		},
	}

	prompt := buildPlanPatchUserPrompt(req)

	// Check required components (QH-001 requirements)
	assert.True(t, strings.Contains(prompt, "create a new feature"), "should contain user input")
	assert.True(t, strings.Contains(prompt, "task-1"), "should contain existing task ID")
	assert.True(t, strings.Contains(prompt, "RUNNING"), "should contain task status")
	assert.True(t, strings.Contains(prompt, "Design"), "should contain phase")
	assert.True(t, strings.Contains(prompt, "WBS Structure"), "should contain WBS section")
	assert.True(t, strings.Contains(prompt, "Conversation History"), "should contain history section")
}
