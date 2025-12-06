package orchestrator

import (
	"context"
	"testing"

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
	// For state transitions we don't need real store/scheduler if we don't start the loop or if loop handles nil gracefully
	// BUT NewExecutionOrchestrator requires them.
	// We can use nil for most since we test Start/Pause/Stop logic which locks/unlocks state.
	// However, Start() spins up a goroutine. We should be careful.
	// Let's create minimal valid objects or mocks if possible.

	// Minimal setup
	emitter := new(MockEventEmitter)
	// We mock Emit calls
	emitter.On("Emit", mock.Anything, mock.Anything).Return()

	orch := NewExecutionOrchestrator(nil, nil, nil, nil, emitter)
	// Suppress logging in tests if needed via custom logger, but defaults to stdout is fine

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

	// Stop again (should work, idempotent-ish in my impl? logic says if IDLE return nil)
	err = orch.Stop()
	assert.NoError(t, err)
}

func TestExecutionOrchestrator_Flow(t *testing.T) {
	// TODO: Test actual runLoop logic requires mocking Scheduler, Executor, TaskStore, Queue.
	// This is more involved. For MVP verification of components, StateTransitions is key.
	// Leaving this for detailed regression testing later.
}
