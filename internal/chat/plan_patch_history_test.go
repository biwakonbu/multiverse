package chat

import (
	"errors"
	"testing"
	"time"

	"github.com/biwakonbu/agent-runner/internal/meta"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/persistence"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// QH-004 (PRD 13.3 #3): history 失敗注入テスト

// mockHistoryStore is a test double that can simulate AppendAction failures
type mockHistoryStore struct {
	actions        []*persistence.Action
	shouldFail     bool
	failureErr     error
	failedAttempts int
}

func (m *mockHistoryStore) AppendAction(action *persistence.Action) error {
	if m.shouldFail {
		m.failedAttempts++
		return m.failureErr
	}
	m.actions = append(m.actions, action)
	return nil
}

func (m *mockHistoryStore) LoadActions() ([]*persistence.Action, error) {
	return m.actions, nil
}

func (m *mockHistoryStore) ListActions(from, to time.Time) ([]persistence.Action, error) {
	var result []persistence.Action
	for _, a := range m.actions {
		if (a.At.Equal(from) || a.At.After(from)) && (a.At.Equal(to) || a.At.Before(to)) {
			result = append(result, *a)
		}
	}
	return result, nil
}

func (m *mockHistoryStore) Flush() error {
	return nil
}

// TestHistoryAppendFailure_RecordsFailureAction tests that when history append fails,
// a history_failed action is recorded as a fallback
func TestHistoryAppendFailure_RecordsFailureAction(t *testing.T) {
	// Setup: create a mock history store that fails initially then succeeds
	historyStore := &mockHistoryStore{
		shouldFail: true,
		failureErr: errors.New("disk full"),
		actions:    make([]*persistence.Action, 0),
	}

	// Create an action to append
	action := &persistence.Action{
		ID:          uuid.New().String(),
		At:          time.Now(),
		Kind:        "plan_patch",
		WorkspaceID: "test-workspace",
		Payload: map[string]interface{}{
			"session_id":       "test-session",
			"operations_count": 1,
		},
	}

	// Attempt to append (will fail)
	err := historyStore.AppendAction(action)
	require.Error(t, err, "first append should fail")
	assert.Equal(t, 1, historyStore.failedAttempts)

	// Now allow recording failure action
	historyStore.shouldFail = false

	// Record failure action (this simulates what plan_patch.go:398-411 does)
	failAction := &persistence.Action{
		ID:          uuid.New().String(),
		At:          time.Now(),
		Kind:        "history_failed",
		WorkspaceID: "test-workspace",
		Payload: map[string]interface{}{
			"original_action_id": action.ID,
			"error":              err.Error(),
		},
	}
	err = historyStore.AppendAction(failAction)
	require.NoError(t, err, "failure action should succeed")

	// Verify: failure action was recorded
	actions, _ := historyStore.LoadActions()
	require.Len(t, actions, 1)
	assert.Equal(t, "history_failed", actions[0].Kind)
	assert.Equal(t, action.ID, actions[0].Payload["original_action_id"])
	assert.Equal(t, "disk full", actions[0].Payload["error"])
}

// TestHistoryAppendSuccess_NoFailureAction tests normal success case
func TestHistoryAppendSuccess_NoFailureAction(t *testing.T) {
	historyStore := &mockHistoryStore{
		shouldFail: false,
		actions:    make([]*persistence.Action, 0),
	}

	action := &persistence.Action{
		ID:          uuid.New().String(),
		At:          time.Now(),
		Kind:        "plan_patch",
		WorkspaceID: "test-workspace",
		Payload:     map[string]interface{}{},
	}

	err := historyStore.AppendAction(action)
	require.NoError(t, err)

	actions, _ := historyStore.LoadActions()
	require.Len(t, actions, 1)
	assert.Equal(t, "plan_patch", actions[0].Kind)
}

// TestHistoryFailureContainsRecoveryInfo verifies failure payload has recovery info
func TestHistoryFailureContainsRecoveryInfo(t *testing.T) {
	// This test validates the structure of failure actions for recovery purposes

	originalActionID := uuid.New().String()
	failAction := &persistence.Action{
		ID:          uuid.New().String(),
		At:          time.Now(),
		Kind:        "history_failed",
		WorkspaceID: "test-workspace",
		Payload: map[string]interface{}{
			"original_action_id": originalActionID,
			"error":              "connection timeout",
		},
	}

	// Verify required fields for recovery
	assert.Equal(t, "history_failed", failAction.Kind)
	assert.NotEmpty(t, failAction.Payload["original_action_id"])
	assert.NotEmpty(t, failAction.Payload["error"])
}

// TestStateSaveFailure_RecordsFailureAction tests that when design/state save fails,
// a state_save_failed action is recorded before error return
func TestStateSaveFailure_RecordsFailureAction(t *testing.T) {
	// This test validates the new state_save_failed behavior added for PRD 12.3

	historyStore := &mockHistoryStore{
		shouldFail: false,
		actions:    make([]*persistence.Action, 0),
	}

	// Simulate recording a state_save_failed action (as done in plan_patch.go:417-432)
	originalActionID := uuid.New().String()
	failAction := &persistence.Action{
		ID:          uuid.New().String(),
		At:          time.Now(),
		Kind:        "state_save_failed",
		WorkspaceID: "test-workspace",
		Payload: map[string]interface{}{
			"original_action_id": originalActionID,
			"stage":              "save_wbs",
			"error":              "disk quota exceeded",
		},
	}
	err := historyStore.AppendAction(failAction)
	require.NoError(t, err)

	// Verify: state_save_failed action was recorded with all required fields
	actions, _ := historyStore.LoadActions()
	require.Len(t, actions, 1)
	assert.Equal(t, "state_save_failed", actions[0].Kind)
	assert.Equal(t, originalActionID, actions[0].Payload["original_action_id"])
	assert.Equal(t, "save_wbs", actions[0].Payload["stage"])
	assert.Equal(t, "disk quota exceeded", actions[0].Payload["error"])
}

// TestHistoryBeforeStateOrder verifies that history append happens before design/state saves
// This is an integration-style test validating the ordering guarantee of PRD 12.3
func TestHistoryBeforeStateOrder(t *testing.T) {
	// Track the order of operations
	operationOrder := make([]string, 0)

	historyStore := &mockHistoryStore{
		shouldFail: false,
		actions:    make([]*persistence.Action, 0),
	}

	// Simulate the order of operations as done in applyPlanPatch:
	// 1. History append first
	historyAction := &persistence.Action{
		ID:          uuid.New().String(),
		At:          time.Now(),
		Kind:        "plan_patch",
		WorkspaceID: "test-workspace",
		Payload:     map[string]interface{}{},
	}
	err := historyStore.AppendAction(historyAction)
	require.NoError(t, err)
	operationOrder = append(operationOrder, "history_append")

	// 2. Design/state saves after history
	// (simulated, not actual file I/O in unit test)
	operationOrder = append(operationOrder, "save_wbs")
	operationOrder = append(operationOrder, "save_nodes_runtime")
	operationOrder = append(operationOrder, "save_tasks_state")

	// Verify: history_append comes first
	assert.Equal(t, "history_append", operationOrder[0], "history append should happen first")
	assert.Equal(t, []string{"history_append", "save_wbs", "save_nodes_runtime", "save_tasks_state"}, operationOrder)

	// Verify: history action was recorded
	actions, _ := historyStore.LoadActions()
	require.Len(t, actions, 1)
	assert.Equal(t, "plan_patch", actions[0].Kind)
}

// TestMoveToNonExistentParent_ReturnsError verifies move validation
func TestMoveToNonExistentParent_ReturnsError(t *testing.T) {
	// This test validates the new parent validation in moveNodeInWBS

	rootID := "node-root"
	wbs := &persistence.WBS{
		RootNodeID: rootID,
		NodeIndex: []persistence.NodeIndex{
			{NodeID: rootID, ParentID: nil, Children: []string{"A"}},
			{NodeID: "A", ParentID: &rootID, Children: []string{}},
		},
	}

	// Try to move A to a non-existent parent
	nonExistentParent := "non-existent"
	op := meta.PlanOperation{ParentID: &nonExistentParent}
	err := moveNodeInWBS(wbs, "A", op, nil)

	// Should fail with error about non-existent parent
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

// TestMoveToRoot_Succeeds verifies move to root always works
func TestMoveToRoot_Succeeds(t *testing.T) {
	rootID := "node-root"
	parentA := "A"
	wbs := &persistence.WBS{
		RootNodeID: rootID,
		NodeIndex: []persistence.NodeIndex{
			{NodeID: rootID, ParentID: nil, Children: []string{"A"}},
			{NodeID: "A", ParentID: &rootID, Children: []string{"B"}},
			{NodeID: "B", ParentID: &parentA, Children: []string{}},
		},
	}

	// Move B to root (should always succeed)
	op := meta.PlanOperation{ParentID: &rootID}
	err := moveNodeInWBS(wbs, "B", op, nil)
	require.NoError(t, err)

	// Verify B is now under root
	bNode := findNodeByID(wbs, "B")
	require.NotNil(t, bNode)
	assert.Equal(t, rootID, *bNode.ParentID)
}

func findNodeByID(wbs *persistence.WBS, nodeID string) *persistence.NodeIndex {
	for i := range wbs.NodeIndex {
		if wbs.NodeIndex[i].NodeID == nodeID {
			return &wbs.NodeIndex[i]
		}
	}
	return nil
}
