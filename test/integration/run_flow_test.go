package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/biwakonbu/agent-runner/internal/core"
	"github.com/biwakonbu/agent-runner/internal/meta"
	"github.com/biwakonbu/agent-runner/internal/mock"
	"github.com/biwakonbu/agent-runner/pkg/config"
)

func TestRunFlow_Success(t *testing.T) {
	// 1. Setup Config
	cfg := &config.TaskConfig{
		Task: config.TaskDetails{
			ID:    "integration-task-1",
			Title: "Integration Task",
			Repo:  "/tmp/integration-repo",
			PRD: config.PRDDetails{
				Text: "Implement a login feature.",
			},
		},
		Runner: config.RunnerConfig{
			Worker: config.WorkerConfig{
				Env: map[string]string{"TEST_ENV": "1"},
			},
		},
	}

	// 2. Setup Mocks
	// Meta: Plan -> Run -> Complete
	mockMeta := &mock.MetaClient{
		PlanTaskFunc: func(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error) {
			return &meta.PlanTaskResponse{
				TaskID: "integration-task-1",
				AcceptanceCriteria: []meta.AcceptanceCriterion{
					{ID: "AC-1", Description: "Login works"},
				},
			}, nil
		},
		NextActionFunc: func(ctx context.Context, summary *meta.TaskSummary) (*meta.NextActionResponse, error) {
			if summary.WorkerRunsCount == 0 {
				return &meta.NextActionResponse{
					Decision: meta.Decision{Action: "run_worker"},
					WorkerCall: meta.WorkerCall{
						WorkerType: "codex-cli",
						Prompt:     "Implement login",
					},
				}, nil
			}
			return &meta.NextActionResponse{
				Decision: meta.Decision{Action: "mark_complete"},
			}, nil
		},
		CompletionAssessmentFunc: func(ctx context.Context, summary *meta.TaskSummary) (*meta.CompletionAssessmentResponse, error) {
			return &meta.CompletionAssessmentResponse{
				AllCriteriaSatisfied: true,
				Summary:              "All criteria satisfied",
				ByCriterion: []meta.CriterionResult{
					{ID: "AC-1", Status: "passed", Comment: "Login implemented"},
				},
			}, nil
		},
	}

	// Worker: Success
	mockWorker := &mock.WorkerExecutor{
		StartFunc: func(ctx context.Context) error {
			return nil
		},
		StopFunc: func(ctx context.Context) error {
			return nil
		},
		RunWorkerFunc: func(ctx context.Context, call meta.WorkerCall, env map[string]string) (*core.WorkerRunResult, error) {
			if env["TEST_ENV"] != "1" {
				t.Error("Env var not passed to worker")
			}
			if call.Prompt != "Implement login" {
				t.Errorf("Unexpected prompt: %s", call.Prompt)
			}
			return &core.WorkerRunResult{
				ID:         "run-1",
				StartedAt:  time.Now(),
				FinishedAt: time.Now(),
				ExitCode:   0,
				Summary:    "Worker executed successfully",
			}, nil
		},
	}

	// Note: Capture output
	var capturedTaskCtx *core.TaskContext
	mockNote := &mock.NoteWriter{
		WriteFunc: func(taskCtx *core.TaskContext) error {
			capturedTaskCtx = taskCtx
			return nil
		},
	}

	// 3. Run
	runner := core.NewRunner(cfg, mockMeta, mockWorker, mockNote)
	ctx := context.Background()
	result, err := runner.Run(ctx)

	// 4. Verify
	if err != nil {
		t.Fatalf("Runner failed: %v", err)
	}

	if result.State != core.StateComplete {
		t.Errorf("Expected state COMPLETE, got %s", result.State)
	}

	if len(result.WorkerRuns) != 1 {
		t.Errorf("Expected 1 worker run, got %d", len(result.WorkerRuns))
	}

	if capturedTaskCtx == nil {
		t.Error("NoteWriter was not called")
	} else if capturedTaskCtx.ID != "integration-task-1" {
		t.Errorf("NoteWriter received wrong task ID: %s", capturedTaskCtx.ID)
	}
}

func TestRunFlow_Failure(t *testing.T) {
	// 1. Setup Config
	cfg := &config.TaskConfig{
		Task: config.TaskDetails{
			ID:    "integration-task-fail",
			Title: "Fail Task",
			Repo:  ".",
			PRD:   config.PRDDetails{Text: "Fail me"},
		},
	}

	// 2. Setup Mocks
	// Meta: Plan -> Run (Worker Fails) -> Abort (simulated by Meta seeing failure and aborting, or just generic fail)
	// Let's simulate Meta returning error on NextAction
	mockMeta := &mock.MetaClient{
		PlanTaskFunc: func(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error) {
			return &meta.PlanTaskResponse{TaskID: "fail", AcceptanceCriteria: []meta.AcceptanceCriterion{}}, nil
		},
		NextActionFunc: func(ctx context.Context, summary *meta.TaskSummary) (*meta.NextActionResponse, error) {
			// Simulate API error
			return nil, context.DeadlineExceeded
		},
	}

	mockWorker := &mock.WorkerExecutor{
		StartFunc: func(ctx context.Context) error {
			return nil
		},
		StopFunc: func(ctx context.Context) error {
			return nil
		},
	} // Should not be called
	mockNote := &mock.NoteWriter{
		WriteFunc: func(taskCtx *core.TaskContext) error { return nil },
	}

	// 3. Run
	runner := core.NewRunner(cfg, mockMeta, mockWorker, mockNote)
	result, err := runner.Run(context.Background())

	// 4. Verify
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if result.State != core.StateFailed {
		t.Errorf("Expected state FAILED, got %s", result.State)
	}
}

func TestRunFlow_WorkerStartFailure(t *testing.T) {
	// 1. Setup Config
	cfg := &config.TaskConfig{
		Task: config.TaskDetails{
			ID:    "integration-task-worker-start-fail",
			Title: "Worker Start Failure Task",
			Repo:  "/tmp/worker-start-fail",
			PRD: config.PRDDetails{
				Text: "This task should fail at worker startup",
			},
		},
		Runner: config.RunnerConfig{
			Worker: config.WorkerConfig{
				Env: map[string]string{"TEST_ENV": "1"},
			},
		},
	}

	// 2. Setup Mocks
	// Meta: Plan -> NextAction (run_worker要求)
	mockMeta := &mock.MetaClient{
		PlanTaskFunc: func(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error) {
			return &meta.PlanTaskResponse{
				TaskID: "integration-task-worker-start-fail",
				AcceptanceCriteria: []meta.AcceptanceCriterion{
					{ID: "AC-1", Description: "Should not reach here"},
				},
			}, nil
		},
		NextActionFunc: func(ctx context.Context, summary *meta.TaskSummary) (*meta.NextActionResponse, error) {
			// Worker起動を要求
			return &meta.NextActionResponse{
				Decision: meta.Decision{Action: "run_worker"},
				WorkerCall: meta.WorkerCall{
					WorkerType: "codex-cli",
					Prompt:     "Attempt to run worker",
				},
			}, nil
		},
	}

	// Worker: Start がエラーを返す
	mockWorker := &mock.WorkerExecutor{
		StartFunc: func(ctx context.Context) error {
			return context.DeadlineExceeded // Worker起動失敗をシミュレート
		},
		StopFunc: func(ctx context.Context) error {
			return nil
		},
		RunWorkerFunc: func(ctx context.Context, call meta.WorkerCall, env map[string]string) (*core.WorkerRunResult, error) {
			_ = call
			t.Error("RunWorker should not be called when Start fails")
			return nil, nil
		},
	}

	mockNote := &mock.NoteWriter{
		WriteFunc: func(taskCtx *core.TaskContext) error { return nil },
	}

	// 3. Run
	runner := core.NewRunner(cfg, mockMeta, mockWorker, mockNote)
	result, err := runner.Run(context.Background())

	// 4. Verify
	if err == nil {
		t.Error("Expected error from Worker.Start(), got nil")
	}
	if result.State != core.StateFailed {
		t.Errorf("Expected state FAILED, got %s", result.State)
	}
}

func TestRunFlow_CompletionAssessmentFailed(t *testing.T) {
	// 1. Setup Config
	cfg := &config.TaskConfig{
		Task: config.TaskDetails{
			ID:    "integration-task-validation-fail",
			Title: "Completion Assessment Failed Task",
			Repo:  "/tmp/validation-fail",
			PRD: config.PRDDetails{
				Text: "Task with failing completion assessment",
			},
		},
		Runner: config.RunnerConfig{
			Worker: config.WorkerConfig{
				Env: map[string]string{"TEST_ENV": "1"},
			},
		},
	}

	// 2. Setup Mocks
	// Meta: Plan -> Run -> Validation (Failed)
	mockMeta := &mock.MetaClient{
		PlanTaskFunc: func(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error) {
			return &meta.PlanTaskResponse{
				TaskID: "integration-task-validation-fail",
				AcceptanceCriteria: []meta.AcceptanceCriterion{
					{ID: "AC-1", Description: "Feature implemented"},
					{ID: "AC-2", Description: "Tests passing"},
				},
			}, nil
		},
		NextActionFunc: func(ctx context.Context, summary *meta.TaskSummary) (*meta.NextActionResponse, error) {
			if summary.WorkerRunsCount == 0 {
				// 初回: Worker実行を要求
				return &meta.NextActionResponse{
					Decision: meta.Decision{Action: "run_worker"},
					WorkerCall: meta.WorkerCall{
						WorkerType: "codex-cli",
						Prompt:     "Implement feature",
					},
				}, nil
			}
			// Worker実行後: 完了マーク（VALIDATING状態へ遷移）
			return &meta.NextActionResponse{
				Decision: meta.Decision{Action: "mark_complete"},
			}, nil
		},
		CompletionAssessmentFunc: func(ctx context.Context, summary *meta.TaskSummary) (*meta.CompletionAssessmentResponse, error) {
			// CompletionAssessment で失敗を返す
			return &meta.CompletionAssessmentResponse{
				AllCriteriaSatisfied: false, // タスク失敗
				Summary:              "AC-2 not satisfied: Tests are still failing",
				ByCriterion: []meta.CriterionResult{
					{ID: "AC-1", Status: "passed", Comment: "Feature implemented"},
					{ID: "AC-2", Status: "failed", Comment: "Tests failing"},
				},
			}, nil
		},
	}

	// Worker: Success（実行自体は成功）
	mockWorker := &mock.WorkerExecutor{
		StartFunc: func(ctx context.Context) error {
			return nil
		},
		StopFunc: func(ctx context.Context) error {
			return nil
		},
		RunWorkerFunc: func(ctx context.Context, call meta.WorkerCall, env map[string]string) (*core.WorkerRunResult, error) {
			_ = call
			return &core.WorkerRunResult{
				ExitCode: 0,
				Summary:  "Done",
			}, nil
		},
	}

	mockNote := &mock.NoteWriter{
		WriteFunc: func(taskCtx *core.TaskContext) error { return nil },
	}

	// 3. Run
	runner := core.NewRunner(cfg, mockMeta, mockWorker, mockNote)
	result, err := runner.Run(context.Background())

	// 4. Verify
	// CompletionAssessment でタスク失敗と判定されるべき
	// 注: runner.Run() は AllCriteriaSatisfied=false の場合も err=nil を返す
	if err != nil {
		t.Logf("runner.Run() returned error: %v (state: %s)", err, result.State)
	}
	if result.State != core.StateFailed {
		t.Errorf("Expected state FAILED, got %s", result.State)
	}

	// ACの評価結果が記録されているか確認
	if len(result.AcceptanceCriteria) != 2 {
		t.Errorf("Expected 2 ACs, got %d", len(result.AcceptanceCriteria))
	}

	// ACの記述内容が正しいか確認
	// Note: Passed状態は TaskContext には永続化されなくなったため、ここでは検証できない
	var ac2Found bool
	for _, ac := range result.AcceptanceCriteria {
		if ac == "Tests passing" {
			ac2Found = true
		} else if ac == "Feature implemented" {
			// Found AC-1
		} else {
			t.Errorf("Unexpected AC: %s", ac)
		}
	}
	if !ac2Found {
		t.Error("AC-2 description 'Tests passing' not found in result")
	}
}
