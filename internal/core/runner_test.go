package core_test

import (
	"context"
	"testing"

	"github.com/biwakonbu/agent-runner/internal/core"
	"github.com/biwakonbu/agent-runner/internal/meta"
	"github.com/biwakonbu/agent-runner/internal/mock"
	"github.com/biwakonbu/agent-runner/pkg/config"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestRunner_Properties(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 5 // Reduced for faster execution
	properties := gopter.NewProperties(parameters)

	properties.Property("Runner completes task when Meta signals completion", prop.ForAll(
		func(acCount int, workerRuns int) bool {
			// 1. Setup - simplified generators
			if acCount < 0 {
				acCount = 0
			}
			if acCount > 3 {
				acCount = 3
			}
			if workerRuns < 0 {
				workerRuns = 0
			}
			if workerRuns > 2 {
				workerRuns = 2
			}

			cfg := &config.TaskConfig{
				Task: config.TaskDetails{
					ID:    "test-task",
					Title: "Test Task",
					Repo:  ".",
					PRD: config.PRDDetails{
						Text: "Simple test PRD",
					},
				},
				Runner: config.RunnerConfig{
					Worker: config.WorkerConfig{
						Env: map[string]string{},
					},
				},
			}

			// Mock Meta
			mockMeta := &mock.MetaClient{
				PlanTaskFunc: func(ctx context.Context, prd string) (*meta.PlanTaskResponse, error) {
					acs := []meta.AcceptanceCriterion{}
					for i := 0; i < acCount; i++ {
						acs = append(acs, meta.AcceptanceCriterion{ID: "AC", Description: "desc"})
					}
					return &meta.PlanTaskResponse{
						TaskID:             "test-task",
						AcceptanceCriteria: acs,
					}, nil
				},
				NextActionFunc: func(ctx context.Context, summary *meta.TaskSummary) (*meta.NextActionResponse, error) {
					if summary.WorkerRunsCount < workerRuns {
						return &meta.NextActionResponse{
							Decision: meta.Decision{Action: "run_worker"},
							WorkerCall: meta.WorkerCall{
								WorkerType: "codex-cli",
								Prompt:     "Do work",
							},
						}, nil
					}
					return &meta.NextActionResponse{
						Decision: meta.Decision{Action: "mark_complete"},
					}, nil
				},
			}

			// Mock Worker
			mockWorker := &mock.WorkerExecutor{
				RunWorkerFunc: func(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error) {
					return &core.WorkerRunResult{
						ExitCode: 0,
						Summary:  "Done",
					}, nil
				},
			}

			// Mock Note
			mockNote := &mock.NoteWriter{
				WriteFunc: func(taskCtx *core.TaskContext) error {
					return nil
				},
			}

			runner := core.NewRunner(cfg, mockMeta, mockWorker, mockNote)

			// 2. Execute
			ctx := context.Background()
			resultCtx, err := runner.Run(ctx)

			// 3. Verify
			if err != nil {
				t.Logf("Error: %v", err)
				return false
			}
			if resultCtx.State != core.StateComplete {
				t.Logf("Expected COMPLETE, got %s", resultCtx.State)
				return false
			}
			if len(resultCtx.WorkerRuns) != workerRuns {
				t.Logf("Expected %d worker runs, got %d", workerRuns, len(resultCtx.WorkerRuns))
				return false
			}
			if len(resultCtx.AcceptanceCriteria) != acCount {
				t.Logf("Expected %d ACs, got %d", acCount, len(resultCtx.AcceptanceCriteria))
				return false
			}

			return true
		},
		gen.IntRange(0, 3), // acCount - reduced range
		gen.IntRange(0, 2), // workerRuns - reduced range
	))

	properties.TestingRun(t)
}
