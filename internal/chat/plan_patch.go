package chat

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/biwakonbu/agent-runner/internal/logging"
	"github.com/biwakonbu/agent-runner/internal/meta"
	"github.com/biwakonbu/agent-runner/internal/orchestrator"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/persistence"
	"github.com/google/uuid"
)

type PlanPatchApplyResult struct {
	CreatedTasks   []orchestrator.Task
	UpdatedTasks   []orchestrator.Task
	DeletedTaskIDs []string
	MovedTaskIDs   []string
}

func (h *Handler) buildPlanPatchRequest(sessionID, message string, existingTasks []orchestrator.Task) *meta.PlanPatchRequest {
	taskSummaries := make([]meta.ExistingTaskSummary, len(existingTasks))
	for i, t := range existingTasks {
		taskSummaries[i] = meta.ExistingTaskSummary{
			ID:           t.ID,
			Title:        t.Title,
			Status:       string(t.Status),
			Dependencies: t.Dependencies,
			PhaseName:    t.PhaseName,
			Milestone:    t.Milestone,
			WBSLevel:     t.WBSLevel,
			ParentID:     t.ParentID,
		}
	}

	// 会話履歴を取得（最新10件）
	recentMessages, err := h.SessionStore.GetRecentMessages(sessionID, 10)
	if err != nil {
		recentMessages = []ChatMessage{}
	}

	conversationHistory := make([]meta.ConversationMessage, len(recentMessages))
	for i, m := range recentMessages {
		conversationHistory[i] = meta.ConversationMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	var wbsOverview *meta.WBSOverview
	if h.Repo != nil {
		if wbs, err := h.Repo.Design().LoadWBS(); err == nil {
			overview := meta.WBSOverview{
				RootNodeID: wbs.RootNodeID,
				NodeIndex:  make([]meta.WBSNodeIndex, len(wbs.NodeIndex)),
			}
			for i, n := range wbs.NodeIndex {
				overview.NodeIndex[i] = meta.WBSNodeIndex{
					NodeID:   n.NodeID,
					ParentID: n.ParentID,
					Children: append([]string{}, n.Children...),
				}
			}
			wbsOverview = &overview
		}
	}

	return &meta.PlanPatchRequest{
		UserInput: message,
		Context: meta.PlanPatchContext{
			WorkspacePath:       h.ProjectRoot,
			ExistingTasks:       taskSummaries,
			ExistingWBS:         wbsOverview,
			ConversationHistory: conversationHistory,
		},
	}
}

func (h *Handler) applyPlanPatch(
	ctx context.Context,
	sessionID string,
	resp *meta.PlanPatchResponse,
	existingTaskIDs map[string]struct{},
	existingTasksByID map[string]orchestrator.Task,
) (*PlanPatchApplyResult, error) {
	logger := logging.WithTraceID(h.logger, ctx)
	now := time.Now()

	resolveRef := func(id string, tempToReal map[string]string) string {
		if real, ok := tempToReal[id]; ok {
			return real
		}
		return id
	}

	// 1) Allocate IDs for creates (temp_id -> real UUID).
	tempToReal := make(map[string]string)
	for _, op := range resp.Operations {
		if op.Op != meta.PlanOpCreate {
			continue
		}
		tempID := strings.TrimSpace(op.TempID)
		if tempID == "" {
			return nil, fmt.Errorf("plan_patch create op requires temp_id")
		}
		if _, exists := tempToReal[tempID]; exists {
			return nil, fmt.Errorf("duplicate temp_id in plan_patch: %s", tempID)
		}
		tempToReal[tempID] = uuid.New().String()
	}

	// 2) Build tasks for create ops.
	tasksToCreate := make([]orchestrator.Task, 0, len(tempToReal))
	createdByID := make(map[string]orchestrator.Task, len(tempToReal))
	placementByID := make(map[string]meta.PlanOperation, len(tempToReal))

	for _, op := range resp.Operations {
		if op.Op != meta.PlanOpCreate {
			continue
		}
		realID := tempToReal[op.TempID]

		if op.Title == nil || strings.TrimSpace(*op.Title) == "" {
			return nil, fmt.Errorf("plan_patch create op requires title (temp_id=%s)", op.TempID)
		}
		desc := ""
		if op.Description != nil {
			desc = strings.TrimSpace(*op.Description)
		}

		deps := make([]string, 0, len(op.Dependencies))
		for _, depRef := range op.Dependencies {
			depID := resolveRef(strings.TrimSpace(depRef), tempToReal)
			if depID == "" {
				continue
			}
			if _, ok := existingTaskIDs[depID]; ok {
				deps = append(deps, depID)
				continue
			}
			// Allow deps to tasks created in this patch.
			if _, ok := createdByID[depID]; ok {
				deps = append(deps, depID)
				continue
			}
			if _, ok := tempToReal[depRef]; ok {
				// It was a temp id, but create ops are processed sequentially; accept.
				deps = append(deps, depID)
				continue
			}
			return nil, fmt.Errorf("unresolved dependency: %s (from temp_id=%s)", depRef, op.TempID)
		}

		task := orchestrator.Task{
			ID:                 realID,
			Title:              strings.TrimSpace(*op.Title),
			Description:        desc,
			Status:             orchestrator.TaskStatusPending,
			PoolID:             "default",
			CreatedAt:          now,
			UpdatedAt:          now,
			Dependencies:       deps,
			WBSLevel:           0,
			PhaseName:          "",
			Milestone:          "",
			SourceChatID:       &sessionID,
			AcceptanceCriteria: op.AcceptanceCriteria,
			Runner: &orchestrator.RunnerSpec{
				MaxLoops:   orchestrator.DefaultRunnerMaxLoops,
				WorkerKind: orchestrator.DefaultWorkerKind,
			},
		}

		if op.WBSLevel != nil {
			task.WBSLevel = *op.WBSLevel
		}
		if op.PhaseName != nil {
			task.PhaseName = strings.TrimSpace(*op.PhaseName)
		}
		if op.Milestone != nil {
			task.Milestone = strings.TrimSpace(*op.Milestone)
		}

		if op.SuggestedImpl != nil {
			validatedPaths := h.validateFilePaths(op.SuggestedImpl.FilePaths)
			task.SuggestedImpl = &orchestrator.SuggestedImpl{
				Language:    op.SuggestedImpl.Language,
				FilePaths:   validatedPaths,
				Constraints: op.SuggestedImpl.Constraints,
			}
		}

		if op.ParentID != nil {
			parent := strings.TrimSpace(*op.ParentID)
			if parent != "" {
				task.ParentID = &parent
			}
		}

		tasksToCreate = append(tasksToCreate, task)
		createdByID[task.ID] = task
		existingTaskIDs[task.ID] = struct{}{}
		existingTasksByID[task.ID] = task

		if op.ParentID != nil || op.Position != nil {
			placementByID[task.ID] = op
		}
	}

	// 3) Persist created tasks into design/state and TaskStore.
	if len(tasksToCreate) > 0 {
		if err := h.persistDesignAndState(ctx, sessionID, tasksToCreate, existingTasksByID); err != nil {
			logger.Error("failed to persist design/state for created tasks", slog.Any("error", err))
			return nil, fmt.Errorf("failed to persist design/state: %w", err)
		}

		for i := range tasksToCreate {
			task := tasksToCreate[i]
			if err := h.TaskStore.SaveTask(&task); err != nil {
				return nil, fmt.Errorf("failed to save task %s: %w", task.ID, err)
			}
			if h.events != nil {
				h.events.Emit(orchestrator.EventTaskCreated, orchestrator.TaskCreatedEvent{Task: task})
			}
		}
	}

	result := &PlanPatchApplyResult{
		CreatedTasks: tasksToCreate,
	}

	if h.Repo == nil {
		// Repo が無い場合は TaskStore のみ更新して終了。
		return result, nil
	}

	// Ensure workspace repo is initialized.
	if err := h.Repo.Init(); err != nil {
		return nil, fmt.Errorf("failed to init workspace repo: %w", err)
	}

	// Reload latest state after create.
	wbs, err := h.Repo.Design().LoadWBS()
	if err != nil {
		if os.IsNotExist(err) {
			wbs = &persistence.WBS{
				WBSID:       uuid.New().String(),
				ProjectRoot: h.ProjectRoot,
				CreatedAt:   now,
				UpdatedAt:   now,
				RootNodeID:  "node-root",
				NodeIndex:   []persistence.NodeIndex{},
			}
		} else {
			return nil, fmt.Errorf("failed to load wbs: %w", err)
		}
	}

	nodesRuntime, err := h.Repo.State().LoadNodesRuntime()
	if err != nil {
		return nil, fmt.Errorf("failed to load nodes runtime: %w", err)
	}
	tasksState, err := h.Repo.State().LoadTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks state: %w", err)
	}

	// Track deletions to clean up dependencies later.
	deleted := make(map[string]struct{})

	// Apply update/move/delete in order (create already handled).
	for _, op := range resp.Operations {
		switch op.Op {
		case meta.PlanOpCreate:
			continue
		case meta.PlanOpUpdate:
			taskID := resolveRef(strings.TrimSpace(op.TaskID), tempToReal)
			if taskID == "" {
				return nil, fmt.Errorf("plan_patch update op requires task_id")
			}
			if _, ok := existingTaskIDs[taskID]; !ok {
				return nil, fmt.Errorf("unknown task_id in plan_patch update: %s", taskID)
			}
			updated, err := h.applyUpdateOp(ctx, taskID, op, tempToReal, existingTaskIDs)
			if err != nil {
				return nil, err
			}
			if updated != nil {
				result.UpdatedTasks = append(result.UpdatedTasks, *updated)
			}
		case meta.PlanOpMove:
			taskID := resolveRef(strings.TrimSpace(op.TaskID), tempToReal)
			if taskID == "" {
				return nil, fmt.Errorf("plan_patch move op requires task_id")
			}
			if _, ok := existingTaskIDs[taskID]; !ok {
				return nil, fmt.Errorf("unknown task_id in plan_patch move: %s", taskID)
			}
			if err := moveNodeInWBS(wbs, taskID, op, tempToReal); err != nil {
				return nil, err
			}
			result.MovedTaskIDs = append(result.MovedTaskIDs, taskID)

			// Optional facet updates on move.
			if op.PhaseName != nil || op.Milestone != nil || op.WBSLevel != nil || op.ParentID != nil {
				if _, err := h.applyUpdateOp(ctx, taskID, op, tempToReal, existingTaskIDs); err != nil {
					return nil, err
				}
			}
		case meta.PlanOpDelete:
			taskID := resolveRef(strings.TrimSpace(op.TaskID), tempToReal)
			if taskID == "" {
				return nil, fmt.Errorf("plan_patch delete op requires task_id")
			}
			if _, ok := deleted[taskID]; ok {
				continue
			}
			existsInWBS := false
			for _, n := range wbs.NodeIndex {
				if n.NodeID == taskID {
					existsInWBS = true
					break
				}
			}
			if _, ok := existingTaskIDs[taskID]; !ok && !existsInWBS {
				return nil, fmt.Errorf("unknown task_id in plan_patch delete: %s", taskID)
			}
			deleteIDs, err := collectDeleteIDs(wbs, taskID, op.Cascade)
			if err != nil {
				return nil, err
			}
			for _, id := range deleteIDs {
				if id == wbs.RootNodeID {
					return nil, fmt.Errorf("cannot delete root node: %s", id)
				}
				if isTaskRunning(tasksState, id) {
					return nil, fmt.Errorf("cannot delete running task: %s", id)
				}
				wasActive := false
				if _, ok := existingTaskIDs[id]; ok {
					wasActive = true
					delete(existingTaskIDs, id)
					delete(existingTasksByID, id)
				}
				deleted[id] = struct{}{}
				result.DeletedTaskIDs = append(result.DeletedTaskIDs, id)
				removeTaskState(tasksState, id)
				if wasActive {
					markNodeObsolete(nodesRuntime, id, now)
				}
				// QH-003: Pass cascade flag for proper child reparenting
				removeNodeFromWBS(wbs, id, op.Cascade)
			}
		default:
			return nil, fmt.Errorf("unknown plan_patch op: %s", op.Op)
		}
	}

	// Ensure create placements are reflected in WBS.
	for id, op := range placementByID {
		if _, ok := deleted[id]; ok {
			continue
		}
		if err := moveNodeInWBS(wbs, id, op, tempToReal); err != nil {
			return nil, err
		}
	}

	// Clean up dependencies referencing deleted nodes.
	if len(deleted) > 0 {
		if err := cleanupDeletedDependencies(h.Repo, tasksState, deleted, now); err != nil {
			return nil, err
		}
	}

	// QH-004: History append FIRST (before design/state updates) for recoverability
	// Per PRD 12.3 / docs/design/orchestrator-persistence-v2.md:92
	historyAction := &persistence.Action{
		ID:          uuid.New().String(),
		At:          now,
		Kind:        "plan_patch",
		WorkspaceID: h.WorkspaceID,
		Payload: map[string]interface{}{
			"session_id":         sessionID,
			"operations_count":   len(resp.Operations),
			"created_task_ids":   taskIDs(result.CreatedTasks),
			"updated_task_ids":   taskIDs(result.UpdatedTasks),
			"deleted_task_ids":   append([]string{}, result.DeletedTaskIDs...),
			"moved_task_ids":     append([]string{}, result.MovedTaskIDs...),
			"meta_understanding": resp.Understanding,
		},
	}
	if h.Repo.History() != nil {
		if err := h.Repo.History().AppendAction(historyAction); err != nil {
			logger.Warn("history append failed, recording failure", slog.Any("error", err))
			// PRD 13.3 #1: Record failure as separate action for recovery tracking (案B)
			failAction := &persistence.Action{
				ID:          uuid.New().String(),
				At:          now,
				Kind:        "history_failed",
				WorkspaceID: h.WorkspaceID,
				Payload: map[string]interface{}{
					"original_action_id": historyAction.ID,
					"error":              err.Error(),
				},
			}
			_ = h.Repo.History().AppendAction(failAction)
		}
	}

	// Then save design/state
	// If save fails, record failure action for recovery tracking (PRD 12.3 requirement)
	wbs.UpdatedAt = now
	if err := h.Repo.Design().SaveWBS(wbs); err != nil {
		if h.Repo.History() != nil {
			failAction := &persistence.Action{
				ID:          uuid.New().String(),
				At:          now,
				Kind:        "state_save_failed",
				WorkspaceID: h.WorkspaceID,
				Payload: map[string]interface{}{
					"original_action_id": historyAction.ID,
					"stage":              "save_wbs",
					"error":              err.Error(),
				},
			}
			_ = h.Repo.History().AppendAction(failAction)
		}
		return nil, fmt.Errorf("failed to save wbs: %w", err)
	}
	if err := h.Repo.State().SaveNodesRuntime(nodesRuntime); err != nil {
		if h.Repo.History() != nil {
			failAction := &persistence.Action{
				ID:          uuid.New().String(),
				At:          now,
				Kind:        "state_save_failed",
				WorkspaceID: h.WorkspaceID,
				Payload: map[string]interface{}{
					"original_action_id": historyAction.ID,
					"stage":              "save_nodes_runtime",
					"error":              err.Error(),
				},
			}
			_ = h.Repo.History().AppendAction(failAction)
		}
		return nil, fmt.Errorf("failed to save nodes runtime: %w", err)
	}
	if err := h.Repo.State().SaveTasks(tasksState); err != nil {
		if h.Repo.History() != nil {
			failAction := &persistence.Action{
				ID:          uuid.New().String(),
				At:          now,
				Kind:        "state_save_failed",
				WorkspaceID: h.WorkspaceID,
				Payload: map[string]interface{}{
					"original_action_id": historyAction.ID,
					"stage":              "save_tasks_state",
					"error":              err.Error(),
				},
			}
			_ = h.Repo.History().AppendAction(failAction)
		}
		return nil, fmt.Errorf("failed to save tasks state: %w", err)
	}

	return result, nil
}

func taskIDs(tasks []orchestrator.Task) []string {
	out := make([]string, 0, len(tasks))
	for _, t := range tasks {
		if t.ID != "" {
			out = append(out, t.ID)
		}
	}
	return out
}

func isTaskRunning(tasksState *persistence.TasksState, taskID string) bool {
	for _, t := range tasksState.Tasks {
		if t.TaskID == taskID && orchestrator.TaskStatus(t.Status) == orchestrator.TaskStatusRunning {
			return true
		}
	}
	return false
}

func removeTaskState(tasksState *persistence.TasksState, taskID string) {
	if tasksState == nil || taskID == "" {
		return
	}
	next := tasksState.Tasks[:0]
	for _, t := range tasksState.Tasks {
		if t.TaskID == taskID {
			continue
		}
		next = append(next, t)
	}
	tasksState.Tasks = next
}

func markNodeObsolete(nodesRuntime *persistence.NodesRuntime, nodeID string, now time.Time) {
	if nodesRuntime == nil || nodeID == "" {
		return
	}
	for i := range nodesRuntime.Nodes {
		if nodesRuntime.Nodes[i].NodeID == nodeID {
			nodesRuntime.Nodes[i].Status = string(persistence.NodeRuntimeStatusObsolete)
			nodesRuntime.Nodes[i].Notes = append(nodesRuntime.Nodes[i].Notes, persistence.NodeNote{
				At:   now,
				By:   "chat-handler",
				Text: "marked obsolete by plan_patch",
			})
			return
		}
	}
	nodesRuntime.Nodes = append(nodesRuntime.Nodes, persistence.NodeRuntime{
		NodeID: nodeID,
		Status: string(persistence.NodeRuntimeStatusObsolete),
		Implementation: persistence.NodeImplementation{
			Files:          []string{},
			LastModifiedAt: now,
			LastModifiedBy: "chat-handler",
		},
		Verification: persistence.NodeVerification{Status: "not_tested"},
		Notes: []persistence.NodeNote{
			{At: now, By: "chat-handler", Text: "added obsolete node by plan_patch"},
		},
	})
}

func collectDeleteIDs(wbs *persistence.WBS, rootID string, cascade bool) ([]string, error) {
	if strings.TrimSpace(rootID) == "" {
		return nil, fmt.Errorf("delete target is empty")
	}
	if !cascade {
		return []string{rootID}, nil
	}
	childrenByID := make(map[string][]string, len(wbs.NodeIndex))
	for _, n := range wbs.NodeIndex {
		childrenByID[n.NodeID] = n.Children
	}

	var out []string
	seen := make(map[string]struct{})
	var walk func(id string)
	walk = func(id string) {
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		out = append(out, id)
		for _, child := range childrenByID[id] {
			walk(child)
		}
	}
	walk(rootID)
	return out, nil
}

// removeNodeFromWBS removes a node from WBS.
// QH-003: When cascade=false, reparent children to the deleted node's parent (采用案A: 子を親に繰り上げ)
// Per PRD 12.2, this maintains WBS invariants (no orphan nodes).
func removeNodeFromWBS(wbs *persistence.WBS, nodeID string, cascade bool) {
	if wbs == nil || nodeID == "" {
		return
	}

	// Find the node to delete and get its info
	var nodeToDelete *persistence.NodeIndex
	var nodeIdx int
	for i := range wbs.NodeIndex {
		if wbs.NodeIndex[i].NodeID == nodeID {
			nodeToDelete = &wbs.NodeIndex[i]
			nodeIdx = i
			break
		}
	}
	if nodeToDelete == nil {
		// Node not found in WBS, still try to clean up any references
		for i := range wbs.NodeIndex {
			children := wbs.NodeIndex[i].Children
			next := children[:0]
			for _, c := range children {
				if c != nodeID {
					next = append(next, c)
				}
			}
			wbs.NodeIndex[i].Children = next
		}
		return
	}

	// Get node's parent ID
	parentID := wbs.RootNodeID
	if nodeToDelete.ParentID != nil && *nodeToDelete.ParentID != "" {
		parentID = *nodeToDelete.ParentID
	}

	// cascade=false: reparent children to deleted node's parent
	if !cascade && len(nodeToDelete.Children) > 0 {
		// Find parent and splice children at deleted node's position
		for i := range wbs.NodeIndex {
			if wbs.NodeIndex[i].NodeID == parentID {
				children := wbs.NodeIndex[i].Children
				// Find position of nodeID in parent's children
				pos := -1
				for j, c := range children {
					if c == nodeID {
						pos = j
						break
					}
				}
				if pos >= 0 {
					// Replace nodeID with its children at that position (splice)
					newChildren := make([]string, 0, len(children)-1+len(nodeToDelete.Children))
					newChildren = append(newChildren, children[:pos]...)
					newChildren = append(newChildren, nodeToDelete.Children...)
					newChildren = append(newChildren, children[pos+1:]...)
					wbs.NodeIndex[i].Children = newChildren
				}
				break
			}
		}

		// Update children's parent_id to point to deleted node's parent
		for _, childID := range nodeToDelete.Children {
			for j := range wbs.NodeIndex {
				if wbs.NodeIndex[j].NodeID == childID {
					parentCopy := parentID
					wbs.NodeIndex[j].ParentID = &parentCopy
					break
				}
			}
		}
	} else {
		// cascade=true or no children: just remove from parent's children list
		for i := range wbs.NodeIndex {
			children := wbs.NodeIndex[i].Children
			next := children[:0]
			for _, c := range children {
				if c != nodeID {
					next = append(next, c)
				}
			}
			wbs.NodeIndex[i].Children = next
		}
	}

	// Remove the node entry itself using nodeIdx for efficiency
	_ = nodeIdx // Used above
	nextIndex := wbs.NodeIndex[:0]
	for _, n := range wbs.NodeIndex {
		if n.NodeID == nodeID {
			continue
		}
		nextIndex = append(nextIndex, n)
	}
	wbs.NodeIndex = nextIndex
}

func moveNodeInWBS(wbs *persistence.WBS, nodeID string, op meta.PlanOperation, tempToReal map[string]string) error {
	if wbs == nil {
		return nil
	}
	if nodeID == "" {
		return fmt.Errorf("move node id is empty")
	}

	parentID := wbs.RootNodeID
	if op.ParentID != nil && strings.TrimSpace(*op.ParentID) != "" {
		parentID = strings.TrimSpace(*op.ParentID)
		if real, ok := tempToReal[parentID]; ok {
			parentID = real
		}
	}

	// Rebuild index map each time (small N, avoids pointer invalidation).
	indexPosByID := make(map[string]int, len(wbs.NodeIndex))
	for i := range wbs.NodeIndex {
		indexPosByID[wbs.NodeIndex[i].NodeID] = i
	}

	// Validate parent exists before move (PRD 11.2 invariant: no orphan nodes)
	// Parent must be either: root, existing in WBS, or created in same patch (via tempToReal)
	if parentID != wbs.RootNodeID {
		if _, exists := indexPosByID[parentID]; !exists {
			return fmt.Errorf("move target parent does not exist: %s", parentID)
		}
	}

	ensureNode := func(id string, parent *string) int {
		if pos, ok := indexPosByID[id]; ok {
			return pos
		}
		wbs.NodeIndex = append(wbs.NodeIndex, persistence.NodeIndex{
			NodeID:   id,
			ParentID: parent,
			Children: []string{},
		})
		pos := len(wbs.NodeIndex) - 1
		indexPosByID[id] = pos
		return pos
	}

	// Ensure root exists.
	if wbs.RootNodeID == "" {
		wbs.RootNodeID = "node-root"
	}
	ensureNode(wbs.RootNodeID, nil)

	// Ensure node exists.
	nodePos := ensureNode(nodeID, &parentID)

	// Detach from old parent.
	oldParentID := wbs.RootNodeID
	if wbs.NodeIndex[nodePos].ParentID != nil && *wbs.NodeIndex[nodePos].ParentID != "" {
		oldParentID = *wbs.NodeIndex[nodePos].ParentID
	}
	oldParentPos := ensureNode(oldParentID, nil)
	wbs.NodeIndex[oldParentPos].Children = removeString(wbs.NodeIndex[oldParentPos].Children, nodeID)

	// Attach to new parent (parent is already validated to exist).
	newParentPos := indexPosByID[parentID]
	children := removeString(wbs.NodeIndex[newParentPos].Children, nodeID)
	children = insertWithPosition(children, nodeID, op.Position, tempToReal)
	wbs.NodeIndex[newParentPos].Children = children

	// Update parent pointer.
	wbs.NodeIndex[nodePos].ParentID = &parentID
	return nil
}

func removeString(xs []string, x string) []string {
	if len(xs) == 0 {
		return xs
	}
	out := xs[:0]
	for _, v := range xs {
		if v == x {
			continue
		}
		out = append(out, v)
	}
	return out
}

func insertWithPosition(children []string, nodeID string, pos *meta.WBSPosition, tempToReal map[string]string) []string {
	if pos == nil {
		return append(children, nodeID)
	}
	// Only one of index/before/after is expected.
	if pos.Index != nil {
		i := *pos.Index
		if i < 0 {
			i = 0
		}
		if i > len(children) {
			i = len(children)
		}
		out := append(children[:i], append([]string{nodeID}, children[i:]...)...)
		return out
	}
	before := strings.TrimSpace(pos.Before)
	after := strings.TrimSpace(pos.After)
	if before != "" {
		if real, ok := tempToReal[before]; ok {
			before = real
		}
		if idx := indexOf(children, before); idx >= 0 {
			out := append(children[:idx], append([]string{nodeID}, children[idx:]...)...)
			return out
		}
		return append(children, nodeID)
	}
	if after != "" {
		if real, ok := tempToReal[after]; ok {
			after = real
		}
		if idx := indexOf(children, after); idx >= 0 {
			insertAt := idx + 1
			out := append(children[:insertAt], append([]string{nodeID}, children[insertAt:]...)...)
			return out
		}
		return append(children, nodeID)
	}
	return append(children, nodeID)
}

func indexOf(xs []string, x string) int {
	for i, v := range xs {
		if v == x {
			return i
		}
	}
	return -1
}

func cleanupDeletedDependencies(repo persistence.WorkspaceRepository, tasksState *persistence.TasksState, deleted map[string]struct{}, now time.Time) error {
	if repo == nil || tasksState == nil || len(deleted) == 0 {
		return nil
	}
	for _, ts := range tasksState.Tasks {
		if ts.NodeID == "" {
			continue
		}
		node, err := repo.Design().GetNode(ts.NodeID)
		if err != nil {
			continue
		}
		changed := false
		next := node.Dependencies[:0]
		for _, dep := range node.Dependencies {
			if _, ok := deleted[dep]; ok {
				changed = true
				continue
			}
			next = append(next, dep)
		}
		if !changed {
			continue
		}
		node.Dependencies = next
		node.UpdatedAt = now
		if err := repo.Design().SaveNode(node); err != nil {
			return fmt.Errorf("failed to save node %s during dependency cleanup: %w", node.NodeID, err)
		}
	}
	return nil
}

func (h *Handler) applyUpdateOp(
	_ context.Context,
	taskID string,
	op meta.PlanOperation,
	tempToReal map[string]string,
	knownTaskIDs map[string]struct{},
) (*orchestrator.Task, error) {
	now := time.Now()

	if h.Repo != nil {
		node, err := h.Repo.Design().GetNode(taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to load node %s: %w", taskID, err)
		}

		if op.Title != nil {
			node.Name = strings.TrimSpace(*op.Title)
		}
		if op.Description != nil {
			node.Summary = strings.TrimSpace(*op.Description)
		}
		if op.PhaseName != nil {
			node.PhaseName = strings.TrimSpace(*op.PhaseName)
		}
		if op.Milestone != nil {
			node.Milestone = strings.TrimSpace(*op.Milestone)
		}
		if op.WBSLevel != nil {
			node.WBSLevel = *op.WBSLevel
		}

		if op.AcceptanceCriteria != nil {
			node.AcceptanceCriteria = op.AcceptanceCriteria
		}

		if op.Dependencies != nil {
			deps := make([]string, 0, len(op.Dependencies))
			for _, depRef := range op.Dependencies {
				depID := strings.TrimSpace(depRef)
				if depID == "" {
					continue
				}
				if real, ok := tempToReal[depID]; ok {
					depID = real
				}
				if _, ok := knownTaskIDs[depID]; !ok {
					return nil, fmt.Errorf("unknown dependency id in update: %s (task_id=%s)", depRef, taskID)
				}
				deps = append(deps, depID)
			}
			node.Dependencies = deps
		}

		if op.SuggestedImpl != nil {
			paths := make([]string, 0, len(op.SuggestedImpl.FilePaths))
			for _, p := range op.SuggestedImpl.FilePaths {
				paths = append(paths, strings.TrimSuffix(p, " (New File)"))
			}
			node.SuggestedImpl = persistence.SuggestedImpl{
				Language:    op.SuggestedImpl.Language,
				FilePaths:   paths,
				Constraints: op.SuggestedImpl.Constraints,
			}
		}

		node.UpdatedAt = now
		if err := h.Repo.Design().SaveNode(node); err != nil {
			return nil, fmt.Errorf("failed to save node %s: %w", node.NodeID, err)
		}
	}

	// Best-effort TaskStore sync.
	storeTask, err := h.TaskStore.LoadTask(taskID)
	if err == nil && storeTask != nil {
		if op.Title != nil && strings.TrimSpace(*op.Title) != "" {
			storeTask.Title = strings.TrimSpace(*op.Title)
		}
		if op.Description != nil {
			storeTask.Description = strings.TrimSpace(*op.Description)
		}
		if op.Dependencies != nil {
			deps := make([]string, 0, len(op.Dependencies))
			for _, depRef := range op.Dependencies {
				depID := strings.TrimSpace(depRef)
				if real, ok := tempToReal[depID]; ok {
					depID = real
				}
				if depID != "" {
					deps = append(deps, depID)
				}
			}
			storeTask.Dependencies = deps
		}
		if op.PhaseName != nil {
			storeTask.PhaseName = strings.TrimSpace(*op.PhaseName)
		}
		if op.Milestone != nil {
			storeTask.Milestone = strings.TrimSpace(*op.Milestone)
		}
		if op.WBSLevel != nil {
			storeTask.WBSLevel = *op.WBSLevel
		}
		if op.ParentID != nil && strings.TrimSpace(*op.ParentID) != "" {
			parent := strings.TrimSpace(*op.ParentID)
			if real, ok := tempToReal[parent]; ok {
				parent = real
			}
			storeTask.ParentID = &parent
		}
		if op.AcceptanceCriteria != nil {
			storeTask.AcceptanceCriteria = op.AcceptanceCriteria
		}
		if op.SuggestedImpl != nil {
			// PRD 13.3 #3: 正規化（NodeDesignと同一ルール）
			normalizedPaths := make([]string, 0, len(op.SuggestedImpl.FilePaths))
			for _, p := range op.SuggestedImpl.FilePaths {
				normalizedPaths = append(normalizedPaths, strings.TrimSuffix(p, " (New File)"))
			}
			storeTask.SuggestedImpl = &orchestrator.SuggestedImpl{
				Language:    op.SuggestedImpl.Language,
				FilePaths:   normalizedPaths,
				Constraints: op.SuggestedImpl.Constraints,
			}
		}
		if err := h.TaskStore.SaveTask(storeTask); err != nil {
			return nil, fmt.Errorf("failed to save task store for %s: %w", taskID, err)
		}
		return storeTask, nil
	}

	return nil, nil
}

func (h *Handler) buildPlanPatchResponseContent(resp *meta.PlanPatchResponse, res *PlanPatchApplyResult) string {
	var b strings.Builder
	b.WriteString(resp.Understanding)
	b.WriteString("\n\n")

	if len(res.CreatedTasks) > 0 {
		b.WriteString(fmt.Sprintf("作成: %d 件\n", len(res.CreatedTasks)))
		for _, t := range res.CreatedTasks {
			b.WriteString(fmt.Sprintf("- **%s**: %s\n", t.Title, t.Description))
		}
		b.WriteString("\n")
	}
	if len(res.UpdatedTasks) > 0 {
		b.WriteString(fmt.Sprintf("更新: %d 件\n", len(res.UpdatedTasks)))
		for _, t := range res.UpdatedTasks {
			b.WriteString(fmt.Sprintf("- %s (%s)\n", t.Title, t.ID))
		}
		b.WriteString("\n")
	}
	if len(res.MovedTaskIDs) > 0 {
		b.WriteString(fmt.Sprintf("移動: %d 件\n", len(res.MovedTaskIDs)))
		for _, id := range res.MovedTaskIDs {
			b.WriteString(fmt.Sprintf("- %s\n", id))
		}
		b.WriteString("\n")
	}
	if len(res.DeletedTaskIDs) > 0 {
		b.WriteString(fmt.Sprintf("削除: %d 件\n", len(res.DeletedTaskIDs)))
		for _, id := range res.DeletedTaskIDs {
			b.WriteString(fmt.Sprintf("- %s\n", id))
		}
		b.WriteString("\n")
	}

	if len(resp.PotentialConflicts) > 0 {
		b.WriteString("**注意**: 以下のファイルで潜在的なコンフリクトが検出されました：\n")
		for _, conflict := range resp.PotentialConflicts {
			b.WriteString(fmt.Sprintf("- `%s`: %s\n", conflict.File, conflict.Warning))
		}
	}

	return b.String()
}
