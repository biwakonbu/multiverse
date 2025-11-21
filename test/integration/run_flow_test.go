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
		RunWorkerFunc: func(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error) {
			if env["TEST_ENV"] != "1" {
				t.Error("Env var not passed to worker")
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

	mockWorker := &mock.WorkerExecutor{} // Should not be called
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
