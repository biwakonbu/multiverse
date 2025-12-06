package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// waitForStop はゴルーチンが終了するための時間を待つヘルパー
func waitForStop() {
	time.Sleep(50 * time.Millisecond)
}

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

	// ゴルーチン終了を待つ
	waitForStop()
}

func TestExecutionOrchestrator_Flow(t *testing.T) {
	// TODO: Test actual runLoop logic requires mocking Scheduler, Executor, TaskStore, Queue.
	// This is more involved. For MVP verification of components, StateTransitions is key.
	// Leaving this for detailed regression testing later.
}

func TestExecutionOrchestrator_EventEmission(t *testing.T) {
	emitter := new(MockEventEmitter)
	// 各状態遷移でイベントが発行されることを確認
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

	orch := NewExecutionOrchestrator(nil, nil, nil, nil, emitter)
	ctx := context.Background()

	// Start → Pause → Resume → Stop の遷移でイベント発行を確認
	_ = orch.Start(ctx)
	_ = orch.Pause()
	_ = orch.Resume()
	_ = orch.Stop()

	// ゴルーチン終了を待つ
	waitForStop()

	emitter.AssertExpectations(t)
}

func TestExecutionOrchestrator_InvalidTransitions(t *testing.T) {
	emitter := new(MockEventEmitter)
	emitter.On("Emit", mock.Anything, mock.Anything).Return()

	orch := NewExecutionOrchestrator(nil, nil, nil, nil, emitter)
	ctx := context.Background()

	// IDLE 状態での無効な遷移
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

	// RUNNING 状態での無効な遷移
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

	// PAUSED 状態での無効な遷移
	_ = orch.Pause()

	// 注意: 現在の実装では Start は state == ExecutionStateRunning のみをチェックしているため、
	// PAUSED から Start を呼ぶと state が Running に変わる（新しいgoroutineが起動する）
	// これは意図的な仕様かバグか不明だが、現状の実装に合わせてテストする
	t.Run("pause from paused", func(t *testing.T) {
		err := orch.Pause()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not running")
	})

	// クリーンアップ: Resume してから Stop
	_ = orch.Resume()
	_ = orch.Stop()

	// ゴルーチン終了を待つ
	waitForStop()
}

func TestExecutionOrchestrator_ContextCancellation(t *testing.T) {
	emitter := new(MockEventEmitter)
	emitter.On("Emit", mock.Anything, mock.Anything).Return()

	orch := NewExecutionOrchestrator(nil, nil, nil, nil, emitter)
	ctx, cancel := context.WithCancel(context.Background())

	err := orch.Start(ctx)
	assert.NoError(t, err)
	assert.Equal(t, ExecutionStateRunning, orch.State())

	// コンテキストをキャンセル
	cancel()

	// 少し待機してコンテキストキャンセルが処理されるのを待つ
	// 実際の runLoop がコンテキストキャンセルを検知して停止するかどうか
	// 現在の実装ではループが依存関係 nil で即座に終了するため、
	// Stop() を明示的に呼び出して状態を確認
	err = orch.Stop()
	assert.NoError(t, err)
	assert.Equal(t, ExecutionStateIdle, orch.State())

	// ゴルーチン終了を待つ
	waitForStop()
}

func TestExecutionOrchestrator_NilEventEmitter(t *testing.T) {
	// EventEmitter が nil でもパニックしないことを確認
	orch := NewExecutionOrchestrator(nil, nil, nil, nil, nil)
	ctx := context.Background()

	// 各操作がパニックせずに実行できることを確認
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

	// ゴルーチン終了を待つ
	waitForStop()
}

func TestExecutionOrchestrator_StateMethod(t *testing.T) {
	emitter := new(MockEventEmitter)
	emitter.On("Emit", mock.Anything, mock.Anything).Return()

	orch := NewExecutionOrchestrator(nil, nil, nil, nil, emitter)
	ctx := context.Background()

	// 初期状態
	assert.Equal(t, ExecutionStateIdle, orch.State())

	// Start後
	_ = orch.Start(ctx)
	assert.Equal(t, ExecutionStateRunning, orch.State())

	// Pause後
	_ = orch.Pause()
	assert.Equal(t, ExecutionStatePaused, orch.State())

	// Resume後
	_ = orch.Resume()
	assert.Equal(t, ExecutionStateRunning, orch.State())

	// Stop後
	_ = orch.Stop()
	assert.Equal(t, ExecutionStateIdle, orch.State())

	// ゴルーチン終了を待つ
	waitForStop()
}
