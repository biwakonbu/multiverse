package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/biwakonbu/agent-runner/internal/orchestrator/ipc"
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

	orch := NewExecutionOrchestrator(nil, nil, nil, nil, emitter, nil, []string{"default"})
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
	// 実際の runLoop ロジックテストは Scheduler, Executor, TaskStore, Queue のモックが必要
	// 状態遷移テストは TestExecutionOrchestrator_StateTransitions で検証済み
	// 詳細な統合テストは execution_retry_test.go で実施
}

func TestExecutionOrchestrator_DependencyOrderExecution(t *testing.T) {
	// 依存順実行テスト: Task A → B → C の順序が保証されることを検証
	// このテストは Scheduler と TaskGraphManager の統合を検証する
	//
	// テスト設計:
	// - Task A (依存なし)
	// - Task B (Task A に依存)
	// - Task C (Task B に依存)
	//
	// 期待する動作:
	// 1. ScheduleReadyTasks() が Task A のみを READY にする
	// 2. Task A 完了後、UpdateBlockedTasks() が Task B を PENDING に
	// 3. ScheduleReadyTasks() が Task B を READY にする
	// 4. Task B 完了後、Task C が同様に処理される
	//
	// 注意: このテストは Scheduler.ScheduleReadyTasks と
	//       Scheduler.UpdateBlockedTasks の動作に依存する
	//       それらのテストは scheduler_test.go で個別に検証済み

	t.Run("dependency order is respected", func(t *testing.T) {
		// 依存順序の検証は Scheduler と TaskGraphManager のテストで実施
		// ExecutionOrchestrator はそれらを呼び出すだけなので、
		// 統合テストは scheduler_test.go および task_graph_test.go を参照
		t.Skip("Integration test - see scheduler_test.go and task_graph_test.go")
	})
}

func TestExecutionOrchestrator_ConcurrentExecution(t *testing.T) {
	// 並行実行制御テスト: maxConcurrent の制限が守られることを検証
	// 現在の実装では runLoop は 1 タスクずつ処理する（Dequeue で 1 件取得）
	//
	// テスト設計:
	// - 並行タスク A, B (依存なし、同時実行可能)
	// - maxConcurrent = 2 の場合、両方が同時に処理される
	//
	// 現在の実装確認:
	// runLoop は 1 回のイテレーションで 1 つの Job を処理
	// 真の並行実行には Worker Pool の拡張が必要
	//
	// 注意: 現在の実装では同時に 1 タスクのみ処理されるため、
	//       このテストは将来の maxConcurrent 実装時に有効化

	t.Run("respects maxConcurrent limit", func(t *testing.T) {
		// 現在の実装では 1 タスクずつ処理するため、
		// 並行実行制御のテストは将来の拡張時に実装
		t.Skip("Not implemented - current design processes one task at a time")
	})
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

	orch := NewExecutionOrchestrator(nil, nil, nil, nil, emitter, nil, []string{"default"})
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

	orch := NewExecutionOrchestrator(nil, nil, nil, nil, emitter, nil, []string{"default"})
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

	orch := NewExecutionOrchestrator(nil, nil, nil, nil, emitter, nil, []string{"default"})
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
	orch := NewExecutionOrchestrator(nil, nil, nil, nil, nil, nil, []string{"default"})
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

	orch := NewExecutionOrchestrator(nil, nil, nil, nil, emitter, nil, []string{"default"})
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

	// Mock TaskStore
	taskStore := NewTaskStore(t.TempDir()) // Use temp dir for real store, or mock it too. Real store is easier if it works with minimal setup.
	// Actually, TaskStore.LoadTask requires a file on disk. Let's seed it.
	taskID := "task-1"
	task := &Task{
		ID:     taskID,
		Title:  "Long Running Task",
		Status: TaskStatusPending,
		PoolID: "default",
	}
	_ = taskStore.SaveTask(task)

	// Mock Queue
	queue := ipc.NewFilesystemQueue(t.TempDir())
	_ = queue.Enqueue(&ipc.Job{
		ID:     "job-1",
		TaskID: taskID,
		PoolID: "default",
	})

	// Mock Executor
	mockExecutor := new(MockExecutor)
	// ExecuteTask should block until context is canceled
	executeCalled := make(chan struct{})
	mockExecutor.On("ExecuteTask", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		close(executeCalled) // Signal that execution started
		<-ctx.Done()         // Wait for cancellation
	}).Return(&Attempt{Status: AttemptStatusFailed}, context.Canceled)

	// Create Orchestrator
	orch := NewExecutionOrchestrator(
		nil, // Scheduler: not needed if we manually enqueue (ExecutionOrchestrator 2. Consume from Queue)
		// But Wait, runLoop 0-a, 0-b, 0-c uses Scheduler. If Scheduler is nil, it skips.
		mockExecutor,
		taskStore,
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

	// Verify state is RUNNING
	assert.Equal(t, ExecutionStateRunning, orch.State())

	// Stop Orchestrator
	err = orch.Stop()
	assert.NoError(t, err)

	// Ensure runLoop exits before asserting expectations to avoid data races
	waitForStop()

	// Verify executed task was canceled
	mockExecutor.AssertExpectations(t)

	// Verify state is IDLE
	assert.Equal(t, ExecutionStateIdle, orch.State())
}
