package orchestrator

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/biwakonbu/agent-runner/internal/logging"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/ipc"
)

// ExecutionState represents the state of the execution loop
type ExecutionState string

const (
	ExecutionStateIdle    ExecutionState = "IDLE"
	ExecutionStateRunning ExecutionState = "RUNNING"
	ExecutionStatePaused  ExecutionState = "PAUSED"
)

// ExecutionOrchestrator manages the autonomous execution loop
type ExecutionOrchestrator struct {
	Scheduler    *Scheduler
	Executor     TaskExecutor
	TaskStore    *TaskStore
	Queue        *ipc.FilesystemQueue
	EventEmitter EventEmitter
	BacklogStore *BacklogStore
	RetryPolicy  *RetryPolicy
	PoolIDs      []string

	state   ExecutionState
	stateMu sync.RWMutex

	// Force Stop support
	runningCancel context.CancelFunc
	cancelMu      sync.Mutex

	stopCh   chan struct{}
	resumeCh chan struct{}

	logger *slog.Logger
}

// NewExecutionOrchestrator creates a new ExecutionOrchestrator
func NewExecutionOrchestrator(
	scheduler *Scheduler,
	executor TaskExecutor,
	taskStore *TaskStore,
	queue *ipc.FilesystemQueue,
	eventEmitter EventEmitter,
	backlogStore *BacklogStore,
	poolIDs []string,
) *ExecutionOrchestrator {
	if len(poolIDs) == 0 {
		poolIDs = []string{"default"}
	}
	return &ExecutionOrchestrator{
		Scheduler:    scheduler,
		Executor:     executor,
		TaskStore:    taskStore,
		Queue:        queue,
		EventEmitter: eventEmitter,
		BacklogStore: backlogStore,
		RetryPolicy:  DefaultRetryPolicy(),
		PoolIDs:      poolIDs,
		state:        ExecutionStateIdle,
		stopCh:       nil,
		resumeCh:     make(chan struct{}),
		logger:       logging.WithComponent(slog.Default(), "execution-orchestrator"),
	}
}

// Start starts the execution loop
func (e *ExecutionOrchestrator) Start(ctx context.Context) error {
	e.stateMu.Lock()
	if e.state == ExecutionStateRunning {
		e.stateMu.Unlock()
		return fmt.Errorf("already running")
	}
	oldState := e.state
	// 再スタートに備え stopCh を作り直す
	e.stopCh = make(chan struct{})
	stopCh := e.stopCh // ゴルーチンに渡すローカルコピー
	e.state = ExecutionStateRunning
	e.stateMu.Unlock()

	e.emitStateChange(oldState, ExecutionStateRunning)
	e.logger.Info("execution orchestrator started")

	// Start the loop in a goroutine
	go e.runLoop(ctx, stopCh)

	return nil
}

// Pause pauses the execution loop
func (e *ExecutionOrchestrator) Pause() error {
	e.stateMu.Lock()
	if e.state != ExecutionStateRunning {
		e.stateMu.Unlock()
		return fmt.Errorf("not running")
	}
	oldState := e.state
	e.state = ExecutionStatePaused
	e.stateMu.Unlock()

	e.emitStateChange(oldState, ExecutionStatePaused)
	e.logger.Info("execution orchestrator paused")
	return nil
}

// Resume resumes the execution loop
func (e *ExecutionOrchestrator) Resume() error {
	e.stateMu.Lock()
	if e.state != ExecutionStatePaused {
		e.stateMu.Unlock()
		return fmt.Errorf("not paused")
	}
	oldState := e.state
	e.state = ExecutionStateRunning
	e.stateMu.Unlock()

	e.emitStateChange(oldState, ExecutionStateRunning)
	e.logger.Info("execution orchestrator resumed")
	return nil
}

// Stop stops the execution loop
func (e *ExecutionOrchestrator) Stop() error {
	e.stateMu.Lock()
	if e.state == ExecutionStateIdle {
		e.stateMu.Unlock()
		return nil
	}
	oldState := e.state
	e.state = ExecutionStateIdle
	stopCh := e.stopCh
	e.stopCh = nil
	e.stateMu.Unlock()

	if stopCh != nil {
		close(stopCh) // runLoop を確実に終了させる
	}

	// Cancel currently running task if any
	e.cancelMu.Lock()
	if e.runningCancel != nil {
		e.logger.Info("canceling running task due to stop signal")
		e.runningCancel()
		e.runningCancel = nil
	}
	e.cancelMu.Unlock()

	e.emitStateChange(oldState, ExecutionStateIdle)
	e.logger.Info("execution orchestrator stopped")
	return nil
}

// State returns the current state
func (e *ExecutionOrchestrator) State() ExecutionState {
	e.stateMu.RLock()
	defer e.stateMu.RUnlock()
	return e.state
}

func (e *ExecutionOrchestrator) emitStateChange(oldState, newState ExecutionState) {
	if e.EventEmitter != nil {
		e.EventEmitter.Emit(EventExecutionStateChange, ExecutionStateChangeEvent{
			OldState:  oldState,
			NewState:  newState,
			Timestamp: time.Now(),
		})
	}
}

func (e *ExecutionOrchestrator) runLoop(ctx context.Context, stopCh <-chan struct{}) {
	ticker := time.NewTicker(2 * time.Second) // Poll every 2s
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("context canceled, stopping loop")
			_ = e.Stop()
			return
		case <-stopCh:
			e.logger.Info("stop signal received, stopping loop")
			return
		case <-ticker.C:
			// Check state
			if e.State() != ExecutionStateRunning {
				continue // Skip if paused or idle
			}

			if e.State() == ExecutionStateIdle {
				return
			}

			// 0-a. Reset Retry Tasks (RETRY_WAIT -> PENDING when backoff expired)
			if e.Scheduler != nil {
				if reset, err := e.Scheduler.ResetRetryTasks(); err != nil {
					e.logger.Error("failed to reset retry tasks", slog.Any("error", err))
				} else {
					for _, id := range reset {
						e.emitTaskStateChange(id, TaskStatusRetryWait, TaskStatusPending)
					}
				}
			}

			// 0-b. Update Blocked Tasks (BLOCKED -> PENDING when dependencies satisfied)
			if e.Scheduler != nil {
				if unblocked, err := e.Scheduler.UpdateBlockedTasks(); err != nil {
					e.logger.Error("failed to update blocked tasks", slog.Any("error", err))
				} else {
					for _, id := range unblocked {
						e.emitTaskStateChange(id, TaskStatusBlocked, TaskStatusPending)
					}
				}
			}

			// 0-c. Set BLOCKED status for pending tasks with unsatisfied dependencies
			if e.Scheduler != nil {
				if newlyBlocked, err := e.Scheduler.SetBlockedStatusForPendingWithUnsatisfiedDeps(); err != nil {
					e.logger.Error("failed to set blocked status for pending tasks", slog.Any("error", err))
				} else {
					for _, id := range newlyBlocked {
						e.emitTaskStateChange(id, TaskStatusPending, TaskStatusBlocked)
					}
				}
			}

			// 1. Schedule Ready Tasks
			// This moves tasks from PENDING/BLOCKED -> READY -> QUEUE
			if e.Scheduler != nil {
				if _, err := e.Scheduler.ScheduleReadyTasks(); err != nil {
					e.logger.Error("failed to schedule ready tasks", slog.Any("error", err))
				}
			}

			// 2. Consume from Queue (Simulating Worker Pool resource availability)
			// Process job for each configured pool
			for _, poolID := range e.PoolIDs {
				job, err := e.Queue.Dequeue(poolID)
				if err != nil {
					e.logger.Error("failed to dequeue job", slog.String("pool_id", poolID), slog.Any("error", err))
					continue
				}

				if job != nil {
					e.processJob(ctx, job)
				}
			}
		}
	}
}

func (e *ExecutionOrchestrator) processJob(ctx context.Context, job *ipc.Job) {
	e.logger.Info("processing job", slog.String("job_id", job.ID), slog.String("task_id", job.TaskID))

	// Load Task
	task, err := e.TaskStore.LoadTask(job.TaskID)
	if err != nil {
		e.logger.Error("failed to load task for job", slog.String("task_id", job.TaskID), slog.Any("error", err))
		_ = e.Queue.Complete(job.ID, job.PoolID)
		return
	}

	// Increment attempt count before execution
	task.AttemptCount++
	if err := e.TaskStore.SaveTask(task); err != nil {
		e.logger.Error("failed to save task attempt count", slog.String("task_id", task.ID), slog.Any("error", err))
		_ = e.Queue.Complete(job.ID, job.PoolID)
		return
	}

	// Create cancellable context for this job
	jobCtx, cancel := context.WithCancel(ctx)
	e.cancelMu.Lock()
	e.runningCancel = cancel
	e.cancelMu.Unlock()

	defer func() {
		e.cancelMu.Lock()
		if e.runningCancel != nil {
			// Ensure cancel is called if not already
			cancel()
			e.runningCancel = nil
		}
		e.cancelMu.Unlock()
	}()

	// Execute Task
	oldStatus := task.Status
	attempt, execErr := e.Executor.ExecuteTask(jobCtx, task)

	// Fetch latest status
	latestTask, loadErr := e.TaskStore.LoadTask(job.TaskID)
	if loadErr == nil && latestTask != nil {
		e.emitTaskStateChange(latestTask.ID, oldStatus, latestTask.Status)
	}

	if execErr != nil {
		e.logger.Error("task execution failed", slog.String("task_id", task.ID), slog.Any("error", execErr))

		// Check if it was canceled
		if jobCtx.Err() == context.Canceled {
			e.logger.Info("task execution canceled by user/system", slog.String("task_id", task.ID))
			// Canceled handling -> likely no retry, just mark as CANCELED or FAILED
			// For now, let HandleFailure decide or just log
		}

		// HandleFailure でリトライ/バックログ追加を判断
		if handleErr := e.HandleFailure(task, execErr, task.AttemptCount); handleErr != nil {
			e.logger.Error("failed to handle task failure", slog.String("task_id", task.ID), slog.Any("error", handleErr))
		}
	} else {
		e.logger.Info("task execution succeeded", slog.String("task_id", task.ID), slog.String("status", string(attempt.Status)))
	}

	// Complete Job
	if err := e.Queue.Complete(job.ID, job.PoolID); err != nil {
		e.logger.Error("failed to complete job", slog.String("job_id", job.ID), slog.Any("error", err))
	}
}

// emitTaskStateChange はタスク状態変更イベントを発行する
func (e *ExecutionOrchestrator) emitTaskStateChange(taskID string, oldStatus, newStatus TaskStatus) {
	if e.EventEmitter != nil {
		e.EventEmitter.Emit(EventTaskStateChange, TaskStateChangeEvent{
			TaskID:    taskID,
			OldStatus: oldStatus,
			NewStatus: newStatus,
			Timestamp: time.Now(),
		})
	}
}

// HandleFailure はタスク失敗時の処理を行う
// PRD FR-P3-004 に基づき、リトライまたはバックログ追加を判断する
func (e *ExecutionOrchestrator) HandleFailure(task *Task, execErr error, attemptNum int) error {
	if e.RetryPolicy == nil {
		e.logger.Warn("no retry policy configured, skipping failure handling")
		return nil
	}

	nextAction := e.RetryPolicy.DetermineNextAction(attemptNum)

	switch nextAction {
	case NextActionRetry:
		// リトライをスケジュール (DB更新)
		backoff := e.RetryPolicy.CalculateBackoff(attemptNum)
		nextRetryAt := time.Now().Add(backoff)

		e.logger.Info("scheduling retry (persisted)",
			slog.String("task_id", task.ID),
			slog.Int("attempt", attemptNum),
			slog.Duration("backoff", backoff),
			slog.Time("next_retry_at", nextRetryAt),
		)

		task.Status = TaskStatusRetryWait
		task.AttemptCount = attemptNum // 失敗した今回の回数を保存
		task.NextRetryAt = &nextRetryAt

		if err := e.TaskStore.SaveTask(task); err != nil {
			return fmt.Errorf("failed to save retry state: %w", err)
		}

		e.emitTaskStateChange(task.ID, TaskStatusFailed, TaskStatusRetryWait)
		return nil

	case NextActionBacklog:
		// バックログに追加
		if e.BacklogStore == nil {
			e.logger.Warn("no backlog store configured, cannot add to backlog")
			return nil
		}
		e.logger.Info("adding task to backlog for human review",
			slog.String("task_id", task.ID),
			slog.Int("attempts", attemptNum),
		)
		item := CreateFailureItem(task.ID, task.Title, execErr, attemptNum)
		if err := e.BacklogStore.Add(item); err != nil {
			return fmt.Errorf("failed to add to backlog: %w", err)
		}
		// バックログ追加イベントを発行
		if e.EventEmitter != nil {
			e.EventEmitter.Emit(EventBacklogAdded, item)
		}
		return nil

	case NextActionFail:
		// 失敗としてマーク（既に Executor で実施済み）
		e.logger.Warn("task permanently failed", slog.String("task_id", task.ID))
		return nil

	default:
		return nil
	}
}
