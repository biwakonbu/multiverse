package orchestrator

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/biwakonbu/agent-runner/internal/logging"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/ipc"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/persistence"
)

// Scheduler manages task execution.
type Scheduler struct {
	Repo   persistence.WorkspaceRepository
	Queue  *ipc.FilesystemQueue
	logger *slog.Logger
	events EventEmitter
}

// NewScheduler creates a new Scheduler.
func NewScheduler(repo persistence.WorkspaceRepository, q *ipc.FilesystemQueue, events EventEmitter) *Scheduler {
	return &Scheduler{
		Repo:   repo,
		Queue:  q,
		logger: logging.WithComponent(slog.Default(), "scheduler"),
		events: events,
	}
}

// ScheduleTask schedules a task for execution.
func (s *Scheduler) ScheduleTask(taskID string) error {
	tasksState, err := s.Repo.State().LoadTasks()
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	var task *persistence.TaskState
	// taskIndex is not needed if we iterate again or save whole state
	for i := range tasksState.Tasks {
		if tasksState.Tasks[i].TaskID == taskID {
			task = &tasksState.Tasks[i]
			break
		}
	}
	if task == nil {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// 依存関係をチェック
	if !s.allDependenciesSatisfied(task) {
		if TaskStatus(task.Status) != TaskStatusBlocked {
			oldStatus := TaskStatus(task.Status)
			task.Status = string(TaskStatusBlocked)
			if err := s.Repo.State().SaveTasks(tasksState); err != nil {
				return fmt.Errorf("failed to update task status: %w", err)
			}
			s.emitStateChange(task.TaskID, oldStatus, TaskStatusBlocked)
		}
		return fmt.Errorf("task has unsatisfied dependencies")
	}

	// Update to READY
	oldStatus := TaskStatus(task.Status)
	task.Status = string(TaskStatusReady)
	if err := s.Repo.State().SaveTasks(tasksState); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	s.emitStateChange(task.TaskID, oldStatus, TaskStatusReady)

	// Create a job for the queue
	job := &ipc.Job{
		ID:      fmt.Sprintf("job-%s-%d", task.TaskID, time.Now().UnixNano()),
		TaskID:  task.TaskID,
		PoolID:  "default", // taskState.PoolID missing? Assuming default.
		Payload: map[string]string{"action": "run_task"},
	}

	if err := s.Queue.Enqueue(job); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}
	s.logger.Info("task scheduled", slog.String("task_id", task.TaskID))
	return nil
}

// allDependenciesSatisfied checks if all dependencies (Node-level) are satisfied.
func (s *Scheduler) allDependenciesSatisfied(task *persistence.TaskState) bool {
	// 1. Get NodeDesign for dependencies
	node, err := s.Repo.Design().GetNode(task.NodeID)
	if err != nil {
		// If node design missing, assume no dependencies (safe fallback?) or block (safer).
		// For now, log warning and block.
		s.logger.Warn("failed to load node design for dependency check", slog.String("node_id", task.NodeID))
		return false
	}

	if len(node.Dependencies) == 0 {
		return true
	}

	// 2. Check NodesRuntime for status of dependency nodes
	nodesRuntime, err := s.Repo.State().LoadNodesRuntime()
	if err != nil {
		return false
	}

	completedNodes := make(map[string]bool)
	for _, nr := range nodesRuntime.Nodes {
		if nr.Status == "implemented" || nr.Status == "verified" {
			completedNodes[nr.NodeID] = true
		}
	}

	for _, depNodeID := range node.Dependencies {
		if !completedNodes[depNodeID] {
			return false
		}
	}

	return true
}

// ScheduleReadyTasks schedules all pending tasks that have satisfied dependencies.
func (s *Scheduler) ScheduleReadyTasks() ([]string, error) {
	tasksState, err := s.Repo.State().LoadTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks state: %w", err)
	}

	scheduled := []string{}
	for i := range tasksState.Tasks {
		task := &tasksState.Tasks[i]
		if TaskStatus(task.Status) == TaskStatusPending {
			if s.allDependenciesSatisfied(task) {
				if err := s.ScheduleTask(task.TaskID); err == nil {
					scheduled = append(scheduled, task.TaskID)
				}
			}
		}
	}

	return scheduled, nil
}

// UpdateBlockedTasks は BLOCKED 状態のタスクで依存が満たされたものを PENDING に戻す
func (s *Scheduler) UpdateBlockedTasks() ([]string, error) {
	tasksState, err := s.Repo.State().LoadTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks state: %w", err)
	}

	unblocked := []string{}
	for i := range tasksState.Tasks {
		task := &tasksState.Tasks[i]
		if TaskStatus(task.Status) == TaskStatusBlocked {
			if s.allDependenciesSatisfied(task) {
				oldStatus := TaskStatus(task.Status)
				task.Status = string(TaskStatusPending)
				if err := s.Repo.State().SaveTasks(tasksState); err != nil {
					s.logger.Warn("failed to unblock task",
						slog.String("task_id", task.TaskID),
						slog.Any("error", err),
					)
					continue
				}
				s.emitStateChange(task.TaskID, oldStatus, TaskStatusPending)
				unblocked = append(unblocked, task.TaskID)
				s.logger.Info("task unblocked",
					slog.String("task_id", task.TaskID),
				)
			}
		}
	}

	return unblocked, nil
}

// SetBlockedStatusForPendingWithUnsatisfiedDeps は依存が満たされていない PENDING タスクを BLOCKED に設定する
func (s *Scheduler) SetBlockedStatusForPendingWithUnsatisfiedDeps() ([]string, error) {
	tasksState, err := s.Repo.State().LoadTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks state: %w", err)
	}

	blocked := []string{}
	for i := range tasksState.Tasks {
		task := &tasksState.Tasks[i]
		if TaskStatus(task.Status) == TaskStatusPending {
			if !s.allDependenciesSatisfied(task) {
				oldStatus := TaskStatus(task.Status)
				task.Status = string(TaskStatusBlocked)
				if err := s.Repo.State().SaveTasks(tasksState); err != nil {
					s.logger.Warn("failed to set task to blocked",
						slog.String("task_id", task.TaskID),
						slog.Any("error", err),
					)
					continue
				}
				s.emitStateChange(task.TaskID, oldStatus, TaskStatusBlocked)
				blocked = append(blocked, task.TaskID)
				s.logger.Info("task set to blocked",
					slog.String("task_id", task.TaskID),
				)
			}
		}
	}

	return blocked, nil
}

// ResetRetryTasks checks for tasks in RETRY_WAIT status that are ready to be retried
// (NextRetryAt <= now) and resets them to PENDING.
func (s *Scheduler) ResetRetryTasks() ([]string, error) {
	tasksState, err := s.Repo.State().LoadTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks state: %w", err)
	}

	now := time.Now()
	reset := []string{}

	for i := range tasksState.Tasks {
		task := &tasksState.Tasks[i]
		if TaskStatus(task.Status) == TaskStatusRetryWait {
			// Check retry time from Inputs (Workaround as models.go lacks NextRetryAt)
			// Or just assume if it's in RETRY_WAIT for a while?
			// For now, let's look for "next_retry_at" in Inputs if exists.
			var nextRetryAt time.Time
			if val, ok := task.Inputs["next_retry_at"].(string); ok {
				if t, err := time.Parse(time.RFC3339, val); err == nil {
					nextRetryAt = t
				}
			}

			// If nextRetryAt is zero (not set) or before now, reset it.
			if nextRetryAt.IsZero() || !now.Before(nextRetryAt) {
				oldStatus := TaskStatus(task.Status)
				task.Status = string(TaskStatusPending)
				// Clear next_retry_at
				delete(task.Inputs, "next_retry_at")

				if err := s.Repo.State().SaveTasks(tasksState); err != nil {
					s.logger.Warn("failed to reset retry task",
						slog.String("task_id", task.TaskID),
						slog.Any("error", err),
					)
					continue
				}
				s.emitStateChange(task.TaskID, oldStatus, TaskStatusPending)
				reset = append(reset, task.TaskID)
				s.logger.Info("task reset for retry (wait time elapsed)",
					slog.String("task_id", task.TaskID),
				)
			}
		}
	}

	return reset, nil
}

// emitStateChange emits a task state change event
func (s *Scheduler) emitStateChange(taskID string, oldStatus, newStatus TaskStatus) {
	if s.events != nil {
		s.events.Emit(EventTaskStateChange, TaskStateChangeEvent{
			TaskID:    taskID,
			OldStatus: oldStatus,
			NewStatus: newStatus,
			Timestamp: time.Now(),
		})
	}
}
