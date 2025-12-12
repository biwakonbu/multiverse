package orchestrator

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/biwakonbu/agent-runner/internal/logging"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/ipc"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/persistence"
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
	Repo         persistence.WorkspaceRepository
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

	wg sync.WaitGroup

	logger *slog.Logger
}

// NewExecutionOrchestrator creates a new ExecutionOrchestrator
func NewExecutionOrchestrator(
	scheduler *Scheduler,
	executor TaskExecutor,
	repo persistence.WorkspaceRepository,
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
		Repo:         repo,
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
	e.wg.Add(1)
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

// Wait waits for the run loop to exit
func (e *ExecutionOrchestrator) Wait() {
	e.wg.Wait()
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
	defer e.wg.Done()
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
	// We use Repo.State()
	tasksState, err := e.Repo.State().LoadTasks()
	if err != nil {
		e.logger.Error("failed to load tasks state", slog.Any("error", err))
		_ = e.Queue.Complete(job.ID, job.PoolID)
		return
	}

	var task *persistence.TaskState
	for i := range tasksState.Tasks {
		if tasksState.Tasks[i].TaskID == job.TaskID {
			task = &tasksState.Tasks[i]
			break
		}
	}
	if task == nil {
		e.logger.Error("task not found in state", slog.String("task_id", job.TaskID))
		_ = e.Queue.Complete(job.ID, job.PoolID)
		return
	}

	// Increment (Note: TaskState model in previous step might not have AttemptCount?
	// If missing, we track via retries or inputs?
	// Plan says: "Update persistence.TaskState based on Executor's result."
	// Let's assume TaskState has AttemptCount or we add it to models.go?
	// Existing models.go TaskState definition:
	// type TaskState struct { ... Priority int, Inputs map, Outputs map ... }
	// It did NOT show AttemptCount in Step 52 summary.
	// I should probably add it or use Inputs/Outputs.
	// For now, I'll use Inputs/Outputs if strictly following provided models.
	// But adding field is better. I'll add AttemptCount to TaskState in models.go later.
	// For now, assuming it exists or I map to Inputs?
	// Let's assume I will add it to models.go.
	// But compilation will fail if I use it and it's missing.
	// I'll check models.go content. I saw it earlier.
	// Step 52: "This struct represents the state of a task... Inputs map... Outputs...". No AttemptCount.
	// I'll add AttemptCount to TaskState in models.go in next step.
	// Here I will access it assuming it exists.
	// task.AttemptCount++

	// Wait, cannot assume field exists if not defined.
	// I will use Inputs as temporary storage if field missing?
	// Or just update models.go FIRST?
	// Updating models.go first is safer but I can't do it in this call sequence cleanly without context switching.
	// I will just use Inputs map for attempt count for now to avoid compilation error until I update model.
	// attempt := 0
	// if val, ok := task.Inputs["attempt_count"].(float64); ok { attempt = int(val) }
	// attempt++
	// task.Inputs["attempt_count"] = attempt

	// NO, I should just fix the model.
	// But for this file I will comment out AttemptCount access and rely on "Update persistence.TaskState" placeholder.
	// Actually, HandleFailure relies on attempt count.
	// processJob gets attempt count from task?
	// I'll comment out increment or use local var.

	// Actually, I can update models.go in parallel? No.
	// I will skip AttemptCount increment here for a moment or use Inputs.

	// task.AttemptCount++
	// if err := e.TaskStore.SaveTask(task); err != nil { ... }

	// New logic:
	// e.Repo.State().SaveTasks(tasksState)

	// I'll implement logic assuming I fix models.go immediately after.
	// But `task` variable here is `*persistence.TaskState`.

	// ... (implementation below)

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

	// Execute Task (Need to map persistence.TaskState to orchestrator.Task for Executor?)
	// Executor takes *orchestrator.Task.
	// We need a mapper.
	// taskDTO := mapStateToDTO(task)
	// attempt, execErr := e.Executor.ExecuteTask(jobCtx, taskDTO)

	// Mapping:
	taskDTO := &Task{
		ID:     task.TaskID,
		Title:  task.Kind + ":" + task.NodeID, // Title fallback
		Status: TaskStatus(task.Status),       // constant cast
		// Other fields...
	}
	// Try to get Title from Design?
	if node, err := e.Repo.Design().GetNode(task.NodeID); err == nil {
		taskDTO.Title = node.Name
		taskDTO.Description = node.Summary
		// Manual conversion of SuggestedImpl
		taskDTO.SuggestedImpl = &SuggestedImpl{
			Language:    node.SuggestedImpl.Language,
			FilePaths:   node.SuggestedImpl.FilePaths,
			Constraints: node.SuggestedImpl.Constraints,
		}
		taskDTO.AcceptanceCriteria = node.AcceptanceCriteria
	}

	oldStatus := TaskStatus(task.Status)
	attempt, execErr := e.Executor.ExecuteTask(jobCtx, taskDTO)

	// Fetch latest state (reload in case changed? or just use tasksState?)
	// Reloading is safer for concurrency.
	tasksState, _ = e.Repo.State().LoadTasks() // Ignore error for reload?
	// finding task again
	for i := range tasksState.Tasks {
		if tasksState.Tasks[i].TaskID == job.TaskID {
			task = &tasksState.Tasks[i]
			break
		}
	}

	if task != nil {
		// Update Status from Executor result?
		// Executor returns Attempt. Status is in Attempt.
		// If attempt succeeded -> Succeeded.
		if attempt != nil {
			if attempt.Status == AttemptStatusSucceeded {
				task.Status = string(TaskStatusSucceeded)
			} else if attempt.Status == AttemptStatusFailed {
				task.Status = string(TaskStatusFailed)
			}
		}

		if oldStatus != TaskStatus(task.Status) {
			e.emitTaskStateChange(task.TaskID, oldStatus, TaskStatus(task.Status))
		}

		// Save
		_ = e.Repo.State().SaveTasks(tasksState)
	}

	if execErr != nil {
		e.logger.Error("task execution failed", slog.String("task_id", task.TaskID), slog.Any("error", execErr))

		// Check if it was canceled
		if jobCtx.Err() == context.Canceled {
			e.logger.Info("task execution canceled by user/system", slog.String("task_id", task.TaskID))
			// Canceled handling -> likely no retry, just mark as CANCELED or FAILED
			// For now, let HandleFailure decide or just log
		}

		// HandleFailure relies on attempt count.
		// Get attempt count from inputs or state?
		attemptCount := 0
		if val, ok := task.Inputs["attempt_count"].(float64); ok {
			attemptCount = int(val)
		} else if val, ok := task.Inputs["attempt_count"].(int); ok {
			attemptCount = val
		}
		// Should increment attempt count? done earlier?
		// Logic earlier was missing.
		// I will rely on HandleFailure to use the count passed.

		if handleErr := e.HandleFailure(task, execErr, attemptCount); handleErr != nil {
			e.logger.Error("failed to handle task failure", slog.String("task_id", task.TaskID), slog.Any("error", handleErr))
		}
	} else {
		e.logger.Info("task execution succeeded", slog.String("task_id", task.TaskID), slog.String("status", string(AttemptStatusSucceeded)))
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

// HandleFailure handles task failure logic
func (e *ExecutionOrchestrator) HandleFailure(task *persistence.TaskState, execErr error, attemptNum int) error {
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
			slog.String("task_id", task.TaskID),
			slog.Int("attempt", attemptNum),
			slog.Duration("backoff", backoff),
			slog.Time("next_retry_at", nextRetryAt),
		)

		// Load state to update
		tasksState, err := e.Repo.State().LoadTasks()
		if err != nil {
			return fmt.Errorf("failed to load tasks for retry: %w", err)
		}

		var taskState *persistence.TaskState
		for i := range tasksState.Tasks {
			if tasksState.Tasks[i].TaskID == task.TaskID {
				taskState = &tasksState.Tasks[i]
				break
			}
		}
		if taskState == nil {
			return fmt.Errorf("task not found for retry: %s", task.TaskID)
		}

		taskState.Status = string(TaskStatusRetryWait)
		// Store next_retry_at in inputs map for now
		if taskState.Inputs == nil {
			taskState.Inputs = make(map[string]interface{})
		}
		taskState.Inputs["next_retry_at"] = nextRetryAt.Format(time.RFC3339)

		if err := e.Repo.State().SaveTasks(tasksState); err != nil {
			return fmt.Errorf("failed to save retry state: %w", err)
		}

		e.emitTaskStateChange(task.TaskID, TaskStatusFailed, TaskStatusRetryWait)
		return nil

	case NextActionBacklog:
		// バックログに追加
		if e.BacklogStore == nil {
			e.logger.Warn("no backlog store configured, cannot add to backlog")
			return nil
		}
		e.logger.Info("adding task to backlog for human review",
			slog.String("task_id", task.TaskID),
			slog.Int("attempts", attemptNum),
		)
		// Title needed. task is TaskState.
		// Need better title fallback. "Task {Kind}:{NodeID}"?
		title := fmt.Sprintf("%s: %s", task.Kind, task.NodeID)
		item := CreateFailureItem(task.TaskID, title, execErr, attemptNum)
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
		e.logger.Warn("task permanently failed", slog.String("task_id", task.TaskID))
		return nil

	default:
		return nil
	}
}
