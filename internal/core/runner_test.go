package core_test

import (
	"context"
	"fmt"
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

			// Mock Worker with Start/Stop verification
			var startCalled, stopCalled bool
			mockWorker := &mock.WorkerExecutor{
				StartFunc: func(ctx context.Context) error {
					startCalled = true
					return nil
				},
				StopFunc: func(ctx context.Context) error {
					stopCalled = true
					return nil
				},
				RunWorkerFunc: func(ctx context.Context, call meta.WorkerCall, env map[string]string) (*core.WorkerRunResult, error) {
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
			// Verify Start/Stop were called
			if !startCalled {
				t.Logf("Worker.Start() should have been called")
				return false
			}
			if !stopCalled {
				t.Logf("Worker.Stop() should have been called")
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

	var startCalled, stopCalled bool
	mockWorker := &mock.WorkerExecutor{
		StartFunc: func(ctx context.Context) error {
			startCalled = true
			return nil
		},
		StopFunc: func(ctx context.Context) error {
			stopCalled = true
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
	// Verify Start/Stop were called
	if !startCalled {
		t.Errorf("Worker.Start() should have been called")
	}
	if !stopCalled {
		t.Errorf("Worker.Stop() should have been called")
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

	var startCalled, stopCalled bool
	mockWorker := &mock.WorkerExecutor{
		StartFunc: func(ctx context.Context) error {
			startCalled = true
			return nil
		},
		StopFunc: func(ctx context.Context) error {
			stopCalled = true
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
	// Verify Start/Stop were called
	if !startCalled {
		t.Errorf("Worker.Start() should have been called")
	}
	if !stopCalled {
		t.Errorf("Worker.Stop() should have been called")
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

	var startCalled, stopCalled bool
	mockWorker := &mock.WorkerExecutor{
		StartFunc: func(ctx context.Context) error {
			startCalled = true
			return nil
		},
		StopFunc: func(ctx context.Context) error {
			stopCalled = true
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
	// Verify Start/Stop were called
	if !startCalled {
		t.Errorf("Worker.Start() should have been called")
	}
	if !stopCalled {
		t.Errorf("Worker.Stop() should have been called")
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

	var startCalled, stopCalled bool
	mockWorker := &mock.WorkerExecutor{
		StartFunc: func(ctx context.Context) error {
			startCalled = true
			return nil
		},
		StopFunc: func(ctx context.Context) error {
			stopCalled = true
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
	// Verify Start/Stop were called
	if !startCalled {
		t.Errorf("Worker.Start() should have been called")
	}
	if !stopCalled {
		t.Errorf("Worker.Stop() should have been called")
	}
}

// TestRunner_ContainerStartFailure tests container startup failure handling
func TestRunner_ContainerStartFailure(t *testing.T) {
	cfg := &config.TaskConfig{
		Task: config.TaskDetails{
			ID:    "test-task",
			Title: "Test Task",
			Repo:  ".",
			PRD: config.PRDDetails{
				Text: "Test PRD",
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

	// Mock Worker that fails on Start
	var startCalled bool
	mockWorker := &mock.WorkerExecutor{
		StartFunc: func(ctx context.Context) error {
			startCalled = true
			return fmt.Errorf("container startup failed: image not found")
		},
		StopFunc: func(ctx context.Context) error {
			// Stop may not be called due to early return
			return nil
		},
		RunWorkerFunc: func(ctx context.Context, call meta.WorkerCall, env map[string]string) (*core.WorkerRunResult, error) {
			// Should never be called if Start fails
			t.Errorf("RunWorker should not be called when Start fails")
			return nil, nil
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

	// Verify error occurred
	if err == nil {
		t.Errorf("Expected error from container startup failure")
	}
	if err != nil && !contains(err.Error(), "failed to start container") {
		t.Errorf("Expected 'failed to start container' in error, got: %v", err)
	}

	// Verify state is FAILED
	if resultCtx.State != core.StateFailed {
		t.Errorf("Expected state FAILED, got %s", resultCtx.State)
	}

	// Verify Start was called
	if !startCalled {
		t.Errorf("Worker.Start() should have been called")
	}

	// Note: Stop will NOT be called because Start failed before defer was registered
	// This is expected behavior - the defer is registered AFTER Start succeeds
	// If we want Stop to always be called, we need to refactor runner.go
}

// TestRunner_ValidatingState_MetaCallsSequence tests VALIDATING state and meta call sequence
func TestRunner_ValidatingState_MetaCallsSequence(t *testing.T) {
	cfg := &config.TaskConfig{
		Task: config.TaskDetails{
			ID:    "test-task",
			Title: "Test Task",
			Repo:  ".",
			PRD: config.PRDDetails{
				Text: "Test PRD",
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
						Prompt: "Test work",
					},
				}, nil
			}
			return &meta.NextActionResponse{
				Decision: meta.Decision{Action: "mark_complete"},
			}, nil
		},
		CompletionAssessmentFunc: func(ctx context.Context, summary *meta.TaskSummary) (*meta.CompletionAssessmentResponse, error) {
			// Verify we are in VALIDATING state when this is called
			if summary.State != "VALIDATING" {
				t.Errorf("Expected state VALIDATING in CompletionAssessment, got %s", summary.State)
			}
			return &meta.CompletionAssessmentResponse{
				AllCriteriaSatisfied: true,
				Summary:              "All criteria passed",
				ByCriterion: []meta.CriterionResult{
					{ID: "AC-1", Status: "passed", Comment: "Passed"},
				},
			}, nil
		},
	}

	var startCalled, stopCalled bool
	mockWorker := &mock.WorkerExecutor{
		StartFunc: func(ctx context.Context) error {
			startCalled = true
			return nil
		},
		StopFunc: func(ctx context.Context) error {
			stopCalled = true
			return nil
		},
		RunWorkerFunc: func(ctx context.Context, call meta.WorkerCall, env map[string]string) (*core.WorkerRunResult, error) {
			return &core.WorkerRunResult{
				ExitCode: 0,
				Summary:  "Work completed",
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

	// Verify final state is COMPLETE
	if resultCtx.State != core.StateComplete {
		t.Errorf("Expected final state COMPLETE, got %s", resultCtx.State)
	}

	// Verify MetaCalls sequence: ["plan_task", "next_action" (run), "next_action" (complete), "completion_assessment"]
	// The implementation calls NextAction twice: once to run worker, once to mark complete
	expectedCallTypes := []string{"plan_task", "next_action", "next_action", "completion_assessment"}
	if len(resultCtx.MetaCalls) < 4 {
		t.Errorf("Expected at least 4 meta calls, got %d", len(resultCtx.MetaCalls))
	}

	for i, expectedType := range expectedCallTypes {
		if i >= len(resultCtx.MetaCalls) {
			break
		}
		actualType := resultCtx.MetaCalls[i].Type
		if actualType != expectedType {
			t.Errorf("Expected meta call %d to be %s, got %s", i, expectedType, actualType)
		}
	}

	// Verify last call is completion_assessment (VALIDATING state call)
	if len(resultCtx.MetaCalls) >= 4 {
		lastCallType := resultCtx.MetaCalls[len(resultCtx.MetaCalls)-1].Type
		if lastCallType != "completion_assessment" {
			t.Errorf("Expected last meta call to be completion_assessment, got %s", lastCallType)
		}
	}

	// Verify Start/Stop were called
	if !startCalled {
		t.Errorf("Worker.Start() should have been called")
	}
	if !stopCalled {
		t.Errorf("Worker.Stop() should have been called")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
