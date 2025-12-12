package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/biwakonbu/agent-runner/internal/orchestrator/ipc"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEventEmitter is a mock implementation of EventEmitter
type MockEventEmitter struct {
	mock.Mock
}

func (m *MockEventEmitter) Emit(eventName string, data any) {
	m.Called(eventName, data)
}

func TestExecutionOrchestrator_StateTransitions(t *testing.T) {
	// Setup dependencies
	emitter := new(MockEventEmitter)
	emitter.On("Emit", mock.Anything, mock.Anything).Return()

	orch := NewExecutionOrchestrator(nil, nil, nil, nil, emitter, nil, []string{"default"})
	ctx := context.Background()

	// Initial State
	assert.Equal(t, ExecutionStateIdle, orch.State())

	// Start
	err := orch.Start(ctx)
	assert.NoError(t, err)
	assert.Equal(t, ExecutionStateRunning, orch.State())

	// Start again (should fail)
	err = orch.Start(ctx)
	assert.Error(t, err)

	// Pause
	err = orch.Pause()
	assert.NoError(t, err)
	assert.Equal(t, ExecutionStatePaused, orch.State())

	// Pause again (should fail)
	err = orch.Pause()
	assert.Error(t, err)

	// Resume
	err = orch.Resume()
	assert.NoError(t, err)
	assert.Equal(t, ExecutionStateRunning, orch.State())

	// Stop
	err = orch.Stop()
	assert.NoError(t, err)
	assert.Equal(t, ExecutionStateIdle, orch.State())

	// Stop again (should work)
	err = orch.Stop()
	assert.NoError(t, err)

	orch.Wait()
}

func TestExecutionOrchestrator_Flow(t *testing.T) {
	// See execution_retry_test.go for flow tests
}

func TestExecutionOrchestrator_DependencyOrderExecution(t *testing.T) {
	t.Run("dependency order is respected", func(t *testing.T) {
		t.Skip("Integration test - see scheduler_test.go and task_graph_test.go")
	})
}

func TestExecutionOrchestrator_ConcurrentExecution(t *testing.T) {
	t.Run("respects maxConcurrent limit", func(t *testing.T) {
		t.Skip("Not implemented - current design processes one task at a time")
	})
}

func TestExecutionOrchestrator_EventEmission(t *testing.T) {
	emitter := new(MockEventEmitter)
	// Expect events
	emitter.On("Emit", "execution:stateChange", mock.MatchedBy(func(data any) bool {
		event, ok := data.(ExecutionStateChangeEvent)
		return ok && event.NewState == ExecutionStateRunning
	})).Return().Once()

	emitter.On("Emit", "execution:stateChange", mock.MatchedBy(func(data any) bool {
		event, ok := data.(ExecutionStateChangeEvent)
		return ok && event.NewState == ExecutionStatePaused
	})).Return().Once()

	emitter.On("Emit", "execution:stateChange", mock.MatchedBy(func(data any) bool {
		event, ok := data.(ExecutionStateChangeEvent)
		return ok && event.NewState == ExecutionStateRunning
	})).Return().Once()

	emitter.On("Emit", "execution:stateChange", mock.MatchedBy(func(data any) bool {
		event, ok := data.(ExecutionStateChangeEvent)
		return ok && event.NewState == ExecutionStateIdle
	})).Return().Once()

	orch := NewExecutionOrchestrator(nil, nil, nil, nil, emitter, nil, []string{"default"})
	ctx := context.Background()

	_ = orch.Start(ctx)
	_ = orch.Pause()
	_ = orch.Resume()
	_ = orch.Stop()
	orch.Wait()

	emitter.AssertExpectations(t)
}

func TestExecutionOrchestrator_InvalidTransitions(t *testing.T) {
	emitter := new(MockEventEmitter)
	emitter.On("Emit", mock.Anything, mock.Anything).Return()

	orch := NewExecutionOrchestrator(nil, nil, nil, nil, emitter, nil, []string{"default"})
	ctx := context.Background()

	t.Run("pause from idle", func(t *testing.T) {
		err := orch.Pause()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not running")
	})

	t.Run("resume from idle", func(t *testing.T) {
		err := orch.Resume()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not paused")
	})

	_ = orch.Start(ctx)

	t.Run("start from running", func(t *testing.T) {
		err := orch.Start(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already running")
	})

	t.Run("resume from running", func(t *testing.T) {
		err := orch.Resume()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not paused")
	})

	_ = orch.Pause()

	t.Run("pause from paused", func(t *testing.T) {
		err := orch.Pause()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not running")
	})

	_ = orch.Resume()
	_ = orch.Stop()
	orch.Wait()
}

func TestExecutionOrchestrator_ContextCancellation(t *testing.T) {
	emitter := new(MockEventEmitter)
	emitter.On("Emit", mock.Anything, mock.Anything).Return()

	orch := NewExecutionOrchestrator(nil, nil, nil, nil, emitter, nil, []string{"default"})
	ctx, cancel := context.WithCancel(context.Background())

	err := orch.Start(ctx)
	assert.NoError(t, err)
	assert.Equal(t, ExecutionStateRunning, orch.State())

	cancel()

	err = orch.Stop()
	assert.NoError(t, err)
	assert.Equal(t, ExecutionStateIdle, orch.State())
	orch.Wait()
}

func TestExecutionOrchestrator_NilEventEmitter(t *testing.T) {
	orch := NewExecutionOrchestrator(nil, nil, nil, nil, nil, nil, []string{"default"})
	ctx := context.Background()

	assert.NotPanics(t, func() {
		_ = orch.Start(ctx)
	})
	assert.NotPanics(t, func() {
		_ = orch.Pause()
	})
	assert.NotPanics(t, func() {
		_ = orch.Resume()
	})
	assert.NotPanics(t, func() {
		_ = orch.Stop()
	})
	orch.Wait()
}

func TestExecutionOrchestrator_StateMethod(t *testing.T) {
	emitter := new(MockEventEmitter)
	emitter.On("Emit", mock.Anything, mock.Anything).Return()

	orch := NewExecutionOrchestrator(nil, nil, nil, nil, emitter, nil, []string{"default"})
	ctx := context.Background()

	assert.Equal(t, ExecutionStateIdle, orch.State())
	_ = orch.Start(ctx)
	assert.Equal(t, ExecutionStateRunning, orch.State())
	_ = orch.Pause()
	assert.Equal(t, ExecutionStatePaused, orch.State())
	_ = orch.Resume()
	assert.Equal(t, ExecutionStateRunning, orch.State())
	_ = orch.Stop()
	assert.Equal(t, ExecutionStateIdle, orch.State())
	orch.Wait()
}

// MockExecutor implements TaskExecutor for testing
type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) ExecuteTask(ctx context.Context, task *Task) (*Attempt, error) {
	args := m.Called(ctx, task)
	if attempt, ok := args.Get(0).(*Attempt); ok {
		return attempt, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestExecutionOrchestrator_Stop_CancelsRunningTask(t *testing.T) {
	// Setup dependencies
	emitter := new(MockEventEmitter)
	emitter.On("Emit", mock.Anything, mock.Anything).Return()

	// New Persistence Repo
	repo, queue := setupTestRepo(t)

	taskID := "task-1"
	// Setup Task State
	saveState(t, repo, []persistence.TaskState{
		{
			TaskID:    taskID,
			NodeID:    "node-1",
			Kind:      "test",
			Status:    string(TaskStatusPending),
			CreatedAt: time.Now(),
		},
	}, nil)
	saveDesign(t, repo, []persistence.NodeDesign{{NodeID: "node-1"}})

	// Initial Queue Enqueue
	_ = queue.Enqueue(&ipc.Job{
		ID:     "job-1",
		TaskID: taskID,
		PoolID: "default",
	})

	// Mock Executor
	mockExecutor := new(MockExecutor)
	executeCalled := make(chan struct{})
	mockExecutor.On("ExecuteTask", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		close(executeCalled) // Signal that execution started
		<-ctx.Done()         // Wait for cancellation
	}).Return(&Attempt{Status: AttemptStatusFailed}, context.Canceled)

	// Create Orchestrator
	orch := NewExecutionOrchestrator(
		nil, // Scheduler skipped for this test manual queue usage
		mockExecutor,
		repo,
		queue,
		emitter,
		nil,
		[]string{"default"},
	)

	ctx := context.Background()

	// Start Orchestrator
	err := orch.Start(ctx)
	assert.NoError(t, err)

	// Wait for ExecuteTask to be called
	select {
	case <-executeCalled:
		// Execution started
	case <-time.After(5 * time.Second):
		t.Fatal("ExecuteTask was not called within timeout")
	}

	assert.Equal(t, ExecutionStateRunning, orch.State())

	// Stop Orchestrator
	err = orch.Stop()
	assert.NoError(t, err)

	orch.Wait()

	mockExecutor.AssertExpectations(t)
	assert.Equal(t, ExecutionStateIdle, orch.State())
}
