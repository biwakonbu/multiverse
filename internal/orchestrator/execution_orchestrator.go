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
	Executor     *Executor
	TaskStore    *TaskStore
	Queue        *ipc.FilesystemQueue
	EventEmitter EventEmitter

	state   ExecutionState
	stateMu sync.RWMutex

	stopCh   chan struct{}
	resumeCh chan struct{}

	// For MVP, we use a simple loop ticker or immediate trigger
	// To support Pause/Resume cleanly, maybe just checking state in loop is enough for MVP

	logger *slog.Logger
}

// NewExecutionOrchestrator creates a new ExecutionOrchestrator
func NewExecutionOrchestrator(
	scheduler *Scheduler,
	executor *Executor,
	taskStore *TaskStore,
	queue *ipc.FilesystemQueue,
	eventEmitter EventEmitter,
) *ExecutionOrchestrator {
	return &ExecutionOrchestrator{
		Scheduler:    scheduler,
		Executor:     executor,
		TaskStore:    taskStore,
		Queue:        queue,
		EventEmitter: eventEmitter,
		state:        ExecutionStateIdle,
		stopCh:       make(chan struct{}),
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
	e.state = ExecutionStateRunning
	e.stateMu.Unlock()

	e.emitStateChange(ExecutionStateIdle, ExecutionStateRunning)
	e.logger.Info("execution orchestrator started")

	// Start the loop in a goroutine
	go e.runLoop(ctx)

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
	e.stateMu.Unlock()

	// Signal stop if we had a dedicated channel, but for simple loop checking state is fine
	// For immediate stop, we might want a cancel context, but let's keep it simple for MVP

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

func (e *ExecutionOrchestrator) runLoop(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second) // Poll every 2s
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("context canceled, stopping loop")
			_ = e.Stop()
			return
		case <-ticker.C:
			// Check state
			if e.State() != ExecutionStateRunning {
				continue // Skip if paused or idle (if idle we should probably exit, but Start spins this up)
			}

			if e.State() == ExecutionStateIdle {
				// Should theoretically cycle back to ensure we stop if someone called Stop()
				return
			}

			// 1. Schedule Ready Tasks
			// This moves tasks from PENDING/BLOCKED -> READY -> QUEUE
			if _, err := e.Scheduler.ScheduleReadyTasks(); err != nil {
				e.logger.Error("failed to schedule ready tasks", slog.Any("error", err))
			}

			// 2. Consume from Queue (Simulating Worker Pool resource availability)
			// For MVP we process one at a time per pool or just default pool
			// Check default pool
			job, err := e.Queue.Dequeue("default") // Assuming "default" pool for now
			if err != nil {
				e.logger.Error("failed to dequeue job", slog.Any("error", err))
				continue
			}

			if job != nil {
				e.processJob(ctx, job)
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
		// Mark job as failed/completed?
		_ = e.Queue.Complete(job.ID, job.PoolID)
		return
	}

	// Execute Task
	// Update status via Scheduler or straight here? Executor updates status.
	attempt, err := e.Executor.ExecuteTask(ctx, task)

	// Emit Task State Change
	// Executor updates DB but maybe doesn't emit event?
	// Executor is "dumb", Orchestrator should probably emit.
	// But Executor updates file.
	// Let's emit here for UI updates
	if e.EventEmitter != nil {
		// Fetch latest status
		task, _ = e.TaskStore.LoadTask(job.TaskID)
		e.EventEmitter.Emit(EventTaskStateChange, TaskStateChangeEvent{
			TaskID:    task.ID,
			OldStatus: TaskStatusReady, // Probably
			NewStatus: task.Status,
			Timestamp: time.Now(),
		})
	}

	if err != nil {
		e.logger.Error("task execution failed", slog.String("task_id", task.ID), slog.Any("error", err))
	} else {
		e.logger.Info("task execution succeeded", slog.String("task_id", task.ID), slog.String("status", string(attempt.Status)))
	}

	// Complete Job
	if err := e.Queue.Complete(job.ID, job.PoolID); err != nil {
		e.logger.Error("failed to complete job", slog.String("job_id", job.ID), slog.Any("error", err))
	}
}
