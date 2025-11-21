package core

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/biwakonbu/agent-runner/internal/meta"
	"github.com/biwakonbu/agent-runner/pkg/config"
	"gopkg.in/yaml.v3"
)

// MetaClient interface for interacting with Meta agent
type MetaClient interface {
	PlanTask(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error)
	NextAction(ctx context.Context, taskSummary *meta.TaskSummary) (*meta.NextActionResponse, error)
	CompletionAssessment(ctx context.Context, taskSummary *meta.TaskSummary) (*meta.CompletionAssessmentResponse, error)
}

// WorkerExecutor interface for executing worker tasks
type WorkerExecutor interface {
	RunWorker(ctx context.Context, prompt string, env map[string]string) (*WorkerRunResult, error)
}

// NoteWriter interface for writing task notes
type NoteWriter interface {
	Write(taskCtx *TaskContext) error
}

// Runner orchestrates the task execution
type Runner struct {
	Config *config.TaskConfig
	Meta   MetaClient
	Worker WorkerExecutor
	Note   NoteWriter
	Logger *slog.Logger
}

// NewRunner creates a new Runner instance
func NewRunner(cfg *config.TaskConfig, m MetaClient, w WorkerExecutor, n NoteWriter) *Runner {
	logger := slog.Default()
	return &Runner{
		Config: cfg,
		Meta:   m,
		Worker: w,
		Note:   n,
		Logger: logger,
	}
}

// Run executes the task
func (r *Runner) Run(ctx context.Context) (*TaskContext, error) {
	// 1. Initialize TaskContext
	taskCtx := &TaskContext{
		ID:        r.Config.Task.ID,
		Title:     r.Config.Task.Title,
		RepoPath:  r.Config.Task.Repo,
		State:     StatePending,
		StartedAt: time.Now(),
	}
	if taskCtx.RepoPath == "" {
		taskCtx.RepoPath = "."
	}
	absRepo, err := filepath.Abs(taskCtx.RepoPath)
	if err != nil {
		return taskCtx, fmt.Errorf("failed to resolve repo path: %w", err)
	}
	taskCtx.RepoPath = absRepo

	// Load PRD
	if r.Config.Task.PRD.Text != "" {
		taskCtx.PRDText = r.Config.Task.PRD.Text
	} else if r.Config.Task.PRD.Path != "" {
		content, err := os.ReadFile(r.Config.Task.PRD.Path)
		if err != nil {
			return taskCtx, fmt.Errorf("failed to read PRD file: %w", err)
		}
		taskCtx.PRDText = string(content)
	} else {
		return taskCtx, fmt.Errorf("PRD not specified")
	}

	// 2. Plan Task
	taskCtx.State = StatePlanning

	// Record PlanTask request
	planRequestYAML := fmt.Sprintf("type: plan_task\nversion: 1\npayload:\n  prd: %q", taskCtx.PRDText)

	plan, err := r.Meta.PlanTask(ctx, taskCtx.PRDText)
	if err != nil {
		taskCtx.State = StateFailed
		return taskCtx, fmt.Errorf("planning failed: %w", err)
	}

	// Record PlanTask response
	planRespData := map[string]interface{}{
		"type":    "plan_task",
		"version": 1,
		"payload": plan,
	}
	planRespBytes, _ := yaml.Marshal(planRespData)
	planResponseYAML := string(planRespBytes)

	taskCtx.MetaCalls = append(taskCtx.MetaCalls, MetaCallLog{
		Type:         "plan_task",
		Timestamp:    time.Now(),
		RequestYAML:  planRequestYAML,
		ResponseYAML: planResponseYAML,
	})

	// Map meta.AcceptanceCriterion to core.AcceptanceCriterion
	for _, ac := range plan.AcceptanceCriteria {
		taskCtx.AcceptanceCriteria = append(taskCtx.AcceptanceCriteria, AcceptanceCriterion{
			ID:          ac.ID,
			Description: ac.Description,
			Passed:      false,
		})
	}

	// 3. Execution Loop
	taskCtx.State = StateRunning
	maxLoops := r.Config.Runner.MaxLoops
	if maxLoops <= 0 {
		maxLoops = 10 // Default value
	}
	for i := 0; i < maxLoops; i++ {
		// Prepare summary
		var metaACs []meta.AcceptanceCriterion
		for _, ac := range taskCtx.AcceptanceCriteria {
			metaACs = append(metaACs, meta.AcceptanceCriterion{
				ID:          ac.ID,
				Description: ac.Description,
				Passed:      ac.Passed,
			})
		}
		summary := &meta.TaskSummary{
			Title:              taskCtx.Title,
			State:              string(taskCtx.State),
			AcceptanceCriteria: metaACs,
			WorkerRunsCount:    len(taskCtx.WorkerRuns),
		}

		// Record NextAction request
		summaryBytes, _ := yaml.Marshal(summary)
		nextActionReqYAML := string(summaryBytes)

		action, err := r.Meta.NextAction(ctx, summary)
		if err != nil {
			taskCtx.State = StateFailed
			return taskCtx, fmt.Errorf("next_action failed: %w", err)
		}

		// Record NextAction response
		actionRespData := map[string]interface{}{
			"type":    "next_action",
			"version": 1,
			"payload": action,
		}
		actionRespBytes, _ := yaml.Marshal(actionRespData)
		nextActionRespYAML := string(actionRespBytes)

		taskCtx.MetaCalls = append(taskCtx.MetaCalls, MetaCallLog{
			Type:         "next_action",
			Timestamp:    time.Now(),
			RequestYAML:  nextActionReqYAML,
			ResponseYAML: nextActionRespYAML,
		})

		if action.Decision.Action == "mark_complete" {
			// Transition to VALIDATING state for completion assessment
			taskCtx.State = StateValidating

			// Prepare TaskSummary with WorkerRuns for completion assessment
			var metaWorkerRuns []meta.WorkerRunSummary
			for _, run := range taskCtx.WorkerRuns {
				metaWorkerRuns = append(metaWorkerRuns, meta.WorkerRunSummary{
					ID:       run.ID,
					ExitCode: run.ExitCode,
					Summary:  run.Summary,
				})
			}

			validationSummary := &meta.TaskSummary{
				Title:              taskCtx.Title,
				State:              string(taskCtx.State),
				AcceptanceCriteria: metaACs,
				WorkerRunsCount:    len(taskCtx.WorkerRuns),
				WorkerRuns:         metaWorkerRuns,
			}

			// Record CompletionAssessment request
			validationSummaryBytes, _ := yaml.Marshal(validationSummary)
			assessmentReqYAML := string(validationSummaryBytes)

			// Call CompletionAssessment to evaluate task completion
			assessment, err := r.Meta.CompletionAssessment(ctx, validationSummary)
			if err != nil {
				taskCtx.State = StateFailed
				return taskCtx, fmt.Errorf("completion assessment failed: %w", err)
			}

			// Record CompletionAssessment response
			assessmentRespData := map[string]interface{}{
				"type":    "completion_assessment",
				"version": 1,
				"payload": assessment,
			}
			assessmentRespBytes, _ := yaml.Marshal(assessmentRespData)
			assessmentRespYAML := string(assessmentRespBytes)

			taskCtx.MetaCalls = append(taskCtx.MetaCalls, MetaCallLog{
				Type:         "completion_assessment",
				Timestamp:    time.Now(),
				RequestYAML:  assessmentReqYAML,
				ResponseYAML: assessmentRespYAML,
			})

			// Update AC Passed flags based on assessment results
			for i := range taskCtx.AcceptanceCriteria {
				for _, result := range assessment.ByCriterion {
					if taskCtx.AcceptanceCriteria[i].ID == result.ID {
						taskCtx.AcceptanceCriteria[i].Passed = (result.Status == "passed")
					}
				}
			}

			// Determine final state based on assessment
			if assessment.AllCriteriaSatisfied {
				taskCtx.State = StateComplete
			} else {
				taskCtx.State = StateFailed
			}
			break
		} else if action.Decision.Action == "run_worker" {
			// Execute Worker
			res, err := r.Worker.RunWorker(ctx, action.WorkerCall.Prompt, r.Config.Runner.Worker.Env)
			if err != nil {
				// Worker execution failed (system error), record it but maybe continue?
				// For now, let's record error in result and continue loop, Meta might retry.
				res = &WorkerRunResult{
					StartedAt:  time.Now(),
					FinishedAt: time.Now(),
					Error:      err,
					Summary:    "Worker execution failed: " + err.Error(),
				}
			}
			taskCtx.WorkerRuns = append(taskCtx.WorkerRuns, *res)
		} else {
			// Unknown action or abort
			taskCtx.State = StateFailed
			return taskCtx, fmt.Errorf("unknown or abort action: %s", action.Decision.Action)
		}
	}

	// 4. Run Test Command (if configured and task completed)
	if taskCtx.State == StateComplete && r.Config.Task.Test.Command != "" {
		if err := r.runTestCommand(ctx, taskCtx); err != nil {
			r.Logger.Warn("test command failed", "err", err)
		}
	}

	// 5. Finish
	taskCtx.FinishedAt = time.Now()

	// Write Note
	if err := r.Note.Write(taskCtx); err != nil {
		r.Logger.Warn("failed to write task note", "err", err)
	}

	return taskCtx, nil
}

// runTestCommand executes the test command configured in the task
func (r *Runner) runTestCommand(ctx context.Context, taskCtx *TaskContext) error {
	testCmd := r.Config.Task.Test.Command
	if testCmd == "" {
		return nil
	}

	r.Logger.Info("running test command", "command", testCmd)

	// Determine working directory
	cwd := r.Config.Task.Test.Cwd
	if cwd == "" {
		cwd = taskCtx.RepoPath
	} else if !filepath.IsAbs(cwd) {
		// Resolve relative path from repo root
		cwd = filepath.Join(taskCtx.RepoPath, cwd)
	}

	// Create command with context
	cmd := exec.CommandContext(ctx, "sh", "-c", testCmd)
	cmd.Dir = cwd

	// Capture output
	output, err := cmd.CombinedOutput()

	// Record test result (even on error)
	taskCtx.TestConfig = &r.Config.Task.Test
	taskCtx.TestResult = &TestResult{
		Command:   testCmd,
		RawOutput: string(output),
	}

	if err != nil {
		taskCtx.TestResult.ExitCode = 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			taskCtx.TestResult.ExitCode = exitErr.ExitCode()
		}
		taskCtx.TestResult.Summary = fmt.Sprintf("Test failed with exit code %d", taskCtx.TestResult.ExitCode)
		r.Logger.Info("test command failed", "exit_code", taskCtx.TestResult.ExitCode)
		return err
	}

	taskCtx.TestResult.ExitCode = 0
	taskCtx.TestResult.Summary = "Test passed"
	r.Logger.Info("test command passed")

	return nil
}
