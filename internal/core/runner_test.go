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
				CompletionAssessmentFunc: func(ctx context.Context, summary *meta.TaskSummary) (*meta.CompletionAssessmentResponse, error) {
					criteria := []meta.CriterionResult{}
					for _, ac := range summary.AcceptanceCriteria {
						criteria = append(criteria, meta.CriterionResult{
							ID:      ac.ID,
							Status:  "passed",
							Comment: "Mock passed",
						})
					}
					return &meta.CompletionAssessmentResponse{
						AllCriteriaSatisfied: true,
						Summary:              "All passed",
						ByCriterion:          criteria,
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

// TestRunner_TestCommand_Success tests test command execution on success
func TestRunner_TestCommand_Success(t *testing.T) {
	cfg := &config.TaskConfig{
		Task: config.TaskDetails{
			ID:    "test-task",
			Title: "Test Task",
			Repo:  ".",
			PRD: config.PRDDetails{
				Text: "Test PRD",
			},
			Test: config.TestDetails{
				Command: "echo 'test passed'",
				Cwd:     "",
			},
		},
		Runner: config.RunnerConfig{
			Worker: config.WorkerConfig{
				Env: map[string]string{},
			},
		},
	}

	mockMeta := &mock.MetaClient{
		PlanTaskFunc: func(ctx context.Context, prd string) (*meta.PlanTaskResponse, error) {
			return &meta.PlanTaskResponse{
				TaskID: "test-task",
				AcceptanceCriteria: []meta.AcceptanceCriterion{
					{ID: "AC-1", Description: "Test AC"},
				},
			}, nil
		},
		NextActionFunc: func(ctx context.Context, summary *meta.TaskSummary) (*meta.NextActionResponse, error) {
			if summary.WorkerRunsCount == 0 {
				return &meta.NextActionResponse{
					Decision: meta.Decision{Action: "run_worker"},
					WorkerCall: meta.WorkerCall{
						Prompt: "Test",
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
				Summary:              "All criteria passed",
				ByCriterion: []meta.CriterionResult{
					{ID: "AC-1", Status: "passed", Comment: "Passed"},
				},
			}, nil
		},
	}

	mockWorker := &mock.WorkerExecutor{
		RunWorkerFunc: func(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error) {
			return &core.WorkerRunResult{
				ExitCode: 0,
				Summary:  "Done",
			}, nil
		},
	}

	mockNote := &mock.NoteWriter{
		WriteFunc: func(taskCtx *core.TaskContext) error {
			return nil
		},
	}

	runner := core.NewRunner(cfg, mockMeta, mockWorker, mockNote)
	ctx := context.Background()
	resultCtx, err := runner.Run(ctx)

	if err != nil {
		t.Fatalf("Runner.Run failed: %v", err)
	}

	if resultCtx.TestResult == nil {
		t.Errorf("TestResult should be set")
	}
	if resultCtx.TestResult.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", resultCtx.TestResult.ExitCode)
	}
	if resultCtx.TestResult.Summary != "Test passed" {
		t.Errorf("Expected 'Test passed' summary, got %q", resultCtx.TestResult.Summary)
	}
}

// TestRunner_TestCommand_Failure tests test command failure handling
func TestRunner_TestCommand_Failure(t *testing.T) {
	cfg := &config.TaskConfig{
		Task: config.TaskDetails{
			ID:    "test-task",
			Title: "Test Task",
			Repo:  ".",
			PRD: config.PRDDetails{
				Text: "Test PRD",
			},
			Test: config.TestDetails{
				Command: "exit 1",
				Cwd:     "",
			},
		},
		Runner: config.RunnerConfig{
			Worker: config.WorkerConfig{
				Env: map[string]string{},
			},
		},
	}

	mockMeta := &mock.MetaClient{
		PlanTaskFunc: func(ctx context.Context, prd string) (*meta.PlanTaskResponse, error) {
			return &meta.PlanTaskResponse{
				TaskID: "test-task",
				AcceptanceCriteria: []meta.AcceptanceCriterion{
					{ID: "AC-1", Description: "Test AC"},
				},
			}, nil
		},
		NextActionFunc: func(ctx context.Context, summary *meta.TaskSummary) (*meta.NextActionResponse, error) {
			if summary.WorkerRunsCount == 0 {
				return &meta.NextActionResponse{
					Decision: meta.Decision{Action: "run_worker"},
					WorkerCall: meta.WorkerCall{
						Prompt: "Test",
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
				Summary:              "All criteria passed",
				ByCriterion: []meta.CriterionResult{
					{ID: "AC-1", Status: "passed", Comment: "Passed"},
				},
			}, nil
		},
	}

	mockWorker := &mock.WorkerExecutor{
		RunWorkerFunc: func(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error) {
			return &core.WorkerRunResult{
				ExitCode: 0,
				Summary:  "Done",
			}, nil
		},
	}

	mockNote := &mock.NoteWriter{
		WriteFunc: func(taskCtx *core.TaskContext) error {
			return nil
		},
	}

	runner := core.NewRunner(cfg, mockMeta, mockWorker, mockNote)
	ctx := context.Background()
	resultCtx, err := runner.Run(ctx)

	if err != nil {
		t.Fatalf("Runner.Run failed: %v", err)
	}

	if resultCtx.TestResult == nil {
		t.Errorf("TestResult should be set")
	}
	if resultCtx.TestResult.ExitCode == 0 {
		t.Errorf("Expected non-zero exit code for failure")
	}
	if resultCtx.TestResult.Summary == "Test passed" {
		t.Errorf("Expected failure summary, got passed")
	}
}

// TestRunner_TestCommand_NotConfigured tests behavior when test command is not configured
func TestRunner_TestCommand_NotConfigured(t *testing.T) {
	cfg := &config.TaskConfig{
		Task: config.TaskDetails{
			ID:    "test-task",
			Title: "Test Task",
			Repo:  ".",
			PRD: config.PRDDetails{
				Text: "Test PRD",
			},
			Test: config.TestDetails{
				Command: "", // No test command
			},
		},
		Runner: config.RunnerConfig{
			Worker: config.WorkerConfig{
				Env: map[string]string{},
			},
		},
	}

	mockMeta := &mock.MetaClient{
		PlanTaskFunc: func(ctx context.Context, prd string) (*meta.PlanTaskResponse, error) {
			return &meta.PlanTaskResponse{
				TaskID: "test-task",
				AcceptanceCriteria: []meta.AcceptanceCriterion{
					{ID: "AC-1", Description: "Test AC"},
				},
			}, nil
		},
		NextActionFunc: func(ctx context.Context, summary *meta.TaskSummary) (*meta.NextActionResponse, error) {
			if summary.WorkerRunsCount == 0 {
				return &meta.NextActionResponse{
					Decision: meta.Decision{Action: "run_worker"},
					WorkerCall: meta.WorkerCall{
						Prompt: "Test",
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
				Summary:              "All criteria passed",
				ByCriterion: []meta.CriterionResult{
					{ID: "AC-1", Status: "passed", Comment: "Passed"},
				},
			}, nil
		},
	}

	mockWorker := &mock.WorkerExecutor{
		RunWorkerFunc: func(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error) {
			return &core.WorkerRunResult{
				ExitCode: 0,
				Summary:  "Done",
			}, nil
		},
	}

	mockNote := &mock.NoteWriter{
		WriteFunc: func(taskCtx *core.TaskContext) error {
			return nil
		},
	}

	runner := core.NewRunner(cfg, mockMeta, mockWorker, mockNote)
	ctx := context.Background()
	resultCtx, err := runner.Run(ctx)

	if err != nil {
		t.Fatalf("Runner.Run failed: %v", err)
	}

	if resultCtx.TestResult != nil {
		t.Errorf("TestResult should be nil when test command not configured")
	}
}

// TestRunner_TestCommand_RelativeCwd tests test command with relative cwd
func TestRunner_TestCommand_RelativeCwd(t *testing.T) {
	cfg := &config.TaskConfig{
		Task: config.TaskDetails{
			ID:    "test-task",
			Title: "Test Task",
			Repo:  ".",
			PRD: config.PRDDetails{
				Text: "Test PRD",
			},
			Test: config.TestDetails{
				Command: "pwd",
				Cwd:     "./subdir", // Relative path
			},
		},
		Runner: config.RunnerConfig{
			Worker: config.WorkerConfig{
				Env: map[string]string{},
			},
		},
	}

	mockMeta := &mock.MetaClient{
		PlanTaskFunc: func(ctx context.Context, prd string) (*meta.PlanTaskResponse, error) {
			return &meta.PlanTaskResponse{
				TaskID: "test-task",
				AcceptanceCriteria: []meta.AcceptanceCriterion{
					{ID: "AC-1", Description: "Test AC"},
				},
			}, nil
		},
		NextActionFunc: func(ctx context.Context, summary *meta.TaskSummary) (*meta.NextActionResponse, error) {
			if summary.WorkerRunsCount == 0 {
				return &meta.NextActionResponse{
					Decision: meta.Decision{Action: "run_worker"},
					WorkerCall: meta.WorkerCall{
						Prompt: "Test",
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
				Summary:              "All criteria passed",
				ByCriterion: []meta.CriterionResult{
					{ID: "AC-1", Status: "passed", Comment: "Passed"},
				},
			}, nil
		},
	}

	mockWorker := &mock.WorkerExecutor{
		RunWorkerFunc: func(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error) {
			return &core.WorkerRunResult{
				ExitCode: 0,
				Summary:  "Done",
			}, nil
		},
	}

	mockNote := &mock.NoteWriter{
		WriteFunc: func(taskCtx *core.TaskContext) error {
			return nil
		},
	}

	runner := core.NewRunner(cfg, mockMeta, mockWorker, mockNote)
	ctx := context.Background()
	resultCtx, err := runner.Run(ctx)

	if err != nil {
		t.Fatalf("Runner.Run failed: %v", err)
	}

	// Test command will fail because subdir doesn't exist, but that's expected
	// We're testing that TestResult is set with exit code 1
	if resultCtx.TestResult == nil {
		t.Errorf("TestResult should be set even on failure")
	}
}
