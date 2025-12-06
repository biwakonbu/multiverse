package orchestrator

import (
	"fmt"
	"log/slog"

	"github.com/biwakonbu/agent-runner/internal/logging"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/ipc"
)

// Scheduler manages task execution.
type Scheduler struct {
	TaskStore    *TaskStore
	GraphManager *TaskGraphManager
	Queue        *ipc.FilesystemQueue
	logger       *slog.Logger
}

// NewScheduler creates a new Scheduler.
func NewScheduler(ts *TaskStore, q *ipc.FilesystemQueue) *Scheduler {
	return &Scheduler{
		TaskStore:    ts,
		GraphManager: NewTaskGraphManager(ts),
		Queue:        q,
		logger:       logging.WithComponent(slog.Default(), "scheduler"),
	}
}

// ScheduleTask schedules a task for execution.
// 依存関係がある場合、全ての依存タスクが完了していなければ BLOCKED 状態に設定される
func (s *Scheduler) ScheduleTask(taskID string) error {
	task, err := s.TaskStore.LoadTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to load task: %w", err)
	}

	if task.Status != TaskStatusPending && task.Status != TaskStatusFailed && task.Status != TaskStatusBlocked {
		return fmt.Errorf("task is not in a schedulable state: %s", task.Status)
	}

	// 依存関係をチェック
	if !s.allDependenciesSatisfied(task) {
		// 依存が満たされていない場合は BLOCKED 状態に設定
		if task.Status != TaskStatusBlocked {
			task.Status = TaskStatusBlocked
			if err := s.TaskStore.SaveTask(task); err != nil {
				return fmt.Errorf("failed to update task status to BLOCKED: %w", err)
			}
			s.logger.Info("task blocked due to unsatisfied dependencies",
				slog.String("task_id", task.ID),
				slog.Any("dependencies", task.Dependencies),
			)
		}
		return fmt.Errorf("task has unsatisfied dependencies: %s", task.ID)
	}

	// Update task status to READY
	task.Status = TaskStatusReady
	if err := s.TaskStore.SaveTask(task); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Create a job for the queue
	job := &ipc.Job{
		ID:      fmt.Sprintf("job-%s-%d", task.ID, task.UpdatedAt.UnixNano()),
		TaskID:  task.ID,
		PoolID:  task.PoolID,
		Payload: map[string]string{"action": "run_task"},
	}

	if err := s.Queue.Enqueue(job); err != nil {
		s.logger.Error("failed to enqueue job",
			slog.String("job_id", job.ID),
			slog.String("task_id", task.ID),
			slog.Any("error", err),
		)
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	s.logger.Info("task scheduled",
		slog.String("task_id", task.ID),
		slog.String("job_id", job.ID),
		slog.String("pool_id", task.PoolID),
	)
	return nil
}

// allDependenciesSatisfied は全ての依存タスクが完了しているかをチェックする
func (s *Scheduler) allDependenciesSatisfied(task *Task) bool {
	if len(task.Dependencies) == 0 {
		return true
	}

	completedStatuses := map[TaskStatus]bool{
		TaskStatusSucceeded: true,
		TaskStatusCompleted: true,
		TaskStatusCanceled:  true,
	}

	for _, depID := range task.Dependencies {
		depTask, err := s.TaskStore.LoadTask(depID)
		if err != nil {
			// 依存タスクが見つからない場合は満たされていないとみなす
			return false
		}
		if !completedStatuses[depTask.Status] {
			return false
		}
	}

	return true
}

// ScheduleReadyTasks は実行可能なタスク（依存が満たされている PENDING タスク）を全てスケジュールする
func (s *Scheduler) ScheduleReadyTasks() ([]string, error) {
	readyTasks, err := s.GraphManager.GetReadyTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to get ready tasks: %w", err)
	}

	scheduled := []string{}
	for _, taskID := range readyTasks {
		if err := s.ScheduleTask(taskID); err != nil {
			s.logger.Warn("failed to schedule ready task",
				slog.String("task_id", taskID),
				slog.Any("error", err),
			)
			continue
		}
		scheduled = append(scheduled, taskID)
	}

	if len(scheduled) > 0 {
		s.logger.Info("scheduled ready tasks",
			slog.Int("count", len(scheduled)),
			slog.Any("task_ids", scheduled),
		)
	}

	return scheduled, nil
}

// UpdateBlockedTasks は BLOCKED 状態のタスクで依存が満たされたものを PENDING に戻す
func (s *Scheduler) UpdateBlockedTasks() ([]string, error) {
	blockedTasks, err := s.TaskStore.ListTasksByStatus(TaskStatusBlocked)
	if err != nil {
		return nil, fmt.Errorf("failed to list blocked tasks: %w", err)
	}

	unblocked := []string{}
	for i := range blockedTasks {
		task := &blockedTasks[i]
		if s.allDependenciesSatisfied(task) {
			task.Status = TaskStatusPending
			if err := s.TaskStore.SaveTask(task); err != nil {
				s.logger.Warn("failed to unblock task",
					slog.String("task_id", task.ID),
					slog.Any("error", err),
				)
				continue
			}
			unblocked = append(unblocked, task.ID)
			s.logger.Info("task unblocked",
				slog.String("task_id", task.ID),
			)
		}
	}

	return unblocked, nil
}

// SetBlockedStatusForPendingWithUnsatisfiedDeps は依存が満たされていない PENDING タスクを BLOCKED に設定する
func (s *Scheduler) SetBlockedStatusForPendingWithUnsatisfiedDeps() ([]string, error) {
	pendingTasks, err := s.TaskStore.ListTasksByStatus(TaskStatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending tasks: %w", err)
	}

	blocked := []string{}
	for i := range pendingTasks {
		task := &pendingTasks[i]
		if len(task.Dependencies) > 0 && !s.allDependenciesSatisfied(task) {
			task.Status = TaskStatusBlocked
			if err := s.TaskStore.SaveTask(task); err != nil {
				s.logger.Warn("failed to set task to blocked",
					slog.String("task_id", task.ID),
					slog.Any("error", err),
				)
				continue
			}
			blocked = append(blocked, task.ID)
			s.logger.Info("task set to blocked",
				slog.String("task_id", task.ID),
				slog.Any("dependencies", task.Dependencies),
			)
		}
	}

	return blocked, nil
}
