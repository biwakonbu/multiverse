package core

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/biwakonbu/agent-runner/internal/logging"
	"github.com/biwakonbu/agent-runner/internal/meta"
	"github.com/biwakonbu/agent-runner/internal/tooling"
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
	RunWorker(ctx context.Context, call meta.WorkerCall, env map[string]string) (*WorkerRunResult, error)
	Start(ctx context.Context) error // Start persistent container
	Stop(ctx context.Context) error  // Stop persistent container
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
	start := time.Now()

	// 1. Initialize TaskContext
	taskCtx := &TaskContext{
		ID:        r.Config.Task.ID,
		Title:     r.Config.Task.Title,
		RepoPath:  r.Config.Task.Repo,
		State:     StatePending,
		StartedAt: time.Now(),
	}

	if r.Config.Task.SuggestedImpl != nil {
		taskCtx.SuggestedImpl = r.Config.Task.SuggestedImpl
	}

	// Create logger with trace ID and task context
	logger := logging.WithTraceID(r.Logger, ctx)
	logger = logging.WithComponent(logger, "runner")
	logger.Info("starting task execution",
		slog.String("task_id", taskCtx.ID),
		slog.String("task_title", taskCtx.Title),
		slog.String("state", string(taskCtx.State)),
	)

	if taskCtx.RepoPath == "" {
		taskCtx.RepoPath = "."
	}
	absRepo, err := filepath.Abs(taskCtx.RepoPath)
	if err != nil {
		logger.Error("failed to resolve repo path", slog.Any("error", err))
		return taskCtx, fmt.Errorf("failed to resolve repo path: %w", err)
	}
	taskCtx.RepoPath = absRepo
	logger.Debug("repo path resolved", slog.String("repo_path", absRepo))

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
	logger.Info("state transition", slog.String("from", string(StatePending)), slog.String("to", string(StatePlanning)))

	// Record PlanTask request
	planRequestYAML := fmt.Sprintf("type: plan_task\nversion: 1\npayload:\n  prd: %q", taskCtx.PRDText)

	logger.Info("calling Meta.PlanTask", slog.String("event_type", "meta:thinking"), slog.String("detail", "Planning task..."))
	logger.Debug("PlanTask request", slog.Int("prd_length", len(taskCtx.PRDText)))
	planStart := time.Now()
	plan, err := r.Meta.PlanTask(ctx, taskCtx.PRDText)
	if err != nil {
		logger.Error("PlanTask failed", slog.Any("error", err), logging.LogDuration(planStart))
		taskCtx.State = StateFailed
		return taskCtx, fmt.Errorf("planning failed: %w", err)
	}
	logger.Info("PlanTask completed",
		slog.Int("criteria_count", len(plan.AcceptanceCriteria)),
		logging.LogDuration(planStart),
	)

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

	// Map meta.AcceptanceCriterion to core.AcceptanceCriterion (stored as strings)
	for _, ac := range plan.AcceptanceCriteria {
		// Just store the description for v2 alignment
		taskCtx.AcceptanceCriteria = append(taskCtx.AcceptanceCriteria, ac.Description)
	}

	// 3. Start Container for the task
	taskCtx.State = StateRunning
	logger.Info("state transition", slog.String("from", string(StatePlanning)), slog.String("to", string(StateRunning)))

	// Start persistent container
	logger.Info("starting worker container", slog.String("event_type", "container:starting"))
	containerStart := time.Now()
	if err := r.Worker.Start(ctx); err != nil {
		logger.Error("failed to start container", slog.Any("error", err), logging.LogDuration(containerStart))
		taskCtx.State = StateFailed
		return taskCtx, fmt.Errorf("failed to start container: %w", err)
	}
	logger.Info("worker container started", slog.String("event_type", "container:started"), logging.LogDuration(containerStart))

	// Ensure container is stopped at the end
	defer func() {
		logger.Info("stopping worker container")
		if err := r.Worker.Stop(ctx); err != nil {
			logger.Warn("failed to stop container", slog.Any("error", err))
		} else {
			logger.Info("worker container stopped")
		}
	}()

	// 4. Execution Loop
	maxLoops := r.Config.Runner.MaxLoops
	if maxLoops <= 0 {
		maxLoops = 10 // Default value
	}
	logger.Info("starting execution loop", slog.Int("max_loops", maxLoops))
	var toolSelector *tooling.Selector
	if r.Config.Runner.Tooling != nil {
		toolSelector = tooling.NewSelector(r.Config.Runner.Tooling)
	}

	for i := 0; i < maxLoops; i++ {
		logger.Info("execution loop iteration", slog.Int("loop", i+1), slog.Int("max", maxLoops))
		// Prepare summary
		var metaACs []meta.AcceptanceCriterion
		for idx, desc := range taskCtx.AcceptanceCriteria {
			metaACs = append(metaACs, meta.AcceptanceCriterion{
				ID:          fmt.Sprintf("AC-%d", idx+1),
				Description: desc,
				Passed:      false, // V2 uses validaiton phase for result, simplifying state here
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

		logger.Info("calling Meta.NextAction", slog.String("event_type", "meta:thinking"), slog.String("detail", "Analyzing..."), slog.Int("worker_runs_count", len(taskCtx.WorkerRuns)))
		actionStart := time.Now()
		action, err := r.Meta.NextAction(ctx, summary)
		if err != nil {
			logger.Error("NextAction failed", slog.Any("error", err), logging.LogDuration(actionStart))
			taskCtx.State = StateFailed
			return taskCtx, fmt.Errorf("next_action failed: %w", err)
		}
		logger.Info("NextAction completed",
			slog.String("action", action.Decision.Action),
			logging.LogDuration(actionStart),
		)

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

			// NOTE: We don't update persistent Passed state for []string based ACs
			// V2 relies on AllCriteriaSatisfied for final decision

			// Determine final state based on assessment
			if assessment.AllCriteriaSatisfied {
				taskCtx.State = StateComplete
			} else {
				taskCtx.State = StateFailed
			}
			break
		} else if action.Decision.Action == "run_worker" {
			// Execute Worker
			logger.Info("executing worker", slog.String("event_type", "worker:running"), slog.String("command", action.WorkerCall.Prompt), slog.Int("prompt_length", len(action.WorkerCall.Prompt)))
			logger.Debug("worker prompt", slog.String("prompt", action.WorkerCall.Prompt))
			baseCall := action.WorkerCall
			attempts := 0
			maxAttempts := 1
			forceMode := false

			if toolSelector != nil {
				if forced, ok := toolSelector.ForceCandidate(); ok {
					forceMode = true
					baseCall = applyWorkerCandidate(baseCall, forced)
					logger.Info("worker tooling forced",
						slog.String("tool", forced.Tool),
						slog.String("model", forced.Model),
					)
				} else if cfg, ok := toolSelector.Category(tooling.CategoryWorker); ok && len(cfg.Candidates) > 0 {
					maxAttempts = len(cfg.Candidates)
				}
			}

			for {
				workerCall := baseCall
				var candidate config.ToolCandidate
				usedTooling := false

				if toolSelector != nil && !forceMode {
					if selected, ok := toolSelector.Select(tooling.CategoryWorker); ok {
						candidate = selected
						workerCall = applyWorkerCandidate(baseCall, candidate)
						usedTooling = true
						logger.Info("worker tooling selected",
							slog.String("tool", candidate.Tool),
							slog.String("model", candidate.Model),
						)
					}
				}

				workerStart := time.Now()
				res, err := r.Worker.RunWorker(ctx, workerCall, r.Config.Runner.Worker.Env)
				if err != nil {
					logger.Error("worker execution failed", slog.Any("error", err), logging.LogDuration(workerStart))
					// Worker execution failed (system error), record it but maybe continue?
					// For now, let's record error in result and continue loop, Meta might retry.
					res = &WorkerRunResult{
						StartedAt:  time.Now(),
						FinishedAt: time.Now(),
						Error:      err,
						Summary:    "Worker execution failed: " + err.Error(),
					}
				} else {
					logger.Info("worker execution completed",
						slog.String("event_type", "worker:completed"),
						slog.Int("exit_code", res.ExitCode),
						slog.Int("output_length", len(res.RawOutput)),
						slog.Any("artifacts", res.Artifacts),
						logging.LogDuration(workerStart),
					)
					logger.Debug("worker output", slog.String("output", res.RawOutput))
				}
				taskCtx.WorkerRuns = append(taskCtx.WorkerRuns, *res)

				rateLimited := false
				if err != nil {
					rateLimited = tooling.IsRateLimitError(err)
				} else if res != nil && res.Error != nil {
					rateLimited = tooling.IsRateLimitError(res.Error)
				}
				if usedTooling && !forceMode && rateLimited && toolSelector.ShouldFallbackOnRateLimit(tooling.CategoryWorker) && attempts+1 < maxAttempts {
					toolSelector.MarkRateLimited(tooling.CategoryWorker, candidate, toolSelector.CooldownSec(tooling.CategoryWorker))
					logger.Warn("worker rate limited; retrying with another tooling candidate",
						slog.String("tool", candidate.Tool),
						slog.String("model", candidate.Model),
					)
					attempts++
					continue
				}
				break
			}
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
	logger.Info("task execution finished",
		slog.String("final_state", string(taskCtx.State)),
		slog.Int("worker_runs_count", len(taskCtx.WorkerRuns)),
		slog.Int("meta_calls_count", len(taskCtx.MetaCalls)),
		logging.LogDuration(start),
	)

	// Write Note
	if err := r.Note.Write(taskCtx); err != nil {
		logger.Warn("failed to write task note", slog.Any("error", err))
	} else {
		logger.Info("task note written", slog.String("task_id", taskCtx.ID))
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

func applyWorkerCandidate(call meta.WorkerCall, candidate config.ToolCandidate) meta.WorkerCall {
	updated := call
	if candidate.Tool != "" {
		updated.WorkerType = candidate.Tool
	}
	if candidate.Model != "" {
		updated.Model = candidate.Model
	}
	if candidate.CLIPath != "" {
		updated.CLIPath = candidate.CLIPath
	}
	if len(candidate.Flags) > 0 {
		updated.Flags = append([]string{}, candidate.Flags...)
	}
	if len(candidate.Env) > 0 {
		updated.Env = mergeStringMap(updated.Env, candidate.Env)
	}
	if len(candidate.ToolSpecific) > 0 {
		updated.ToolSpecific = mergeToolSpecific(updated.ToolSpecific, candidate.ToolSpecific)
	}
	return updated
}

func mergeStringMap(base, override map[string]string) map[string]string {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}
	out := make(map[string]string, len(base)+len(override))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		out[k] = v
	}
	return out
}

func mergeToolSpecific(base, override map[string]interface{}) map[string]interface{} {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}
	out := make(map[string]interface{}, len(base)+len(override))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		out[k] = v
	}
	return out
}
