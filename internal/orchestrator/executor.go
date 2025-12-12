package orchestrator

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"github.com/biwakonbu/agent-runner/internal/logging"
	"github.com/google/uuid"
)

// TaskExecutor defines the interface for executing tasks
type TaskExecutor interface {
	ExecuteTask(ctx context.Context, task *Task) (*Attempt, error)
}

// Executor wraps AgentRunner Core execution.
type Executor struct {
	AgentRunnerPath string // Path to agent-runner binary
	ProjectRoot     string // Root directory of the project
	logger          *slog.Logger
	events          EventEmitter // Event emitter for streaming logs
}

// NewExecutor creates a new Executor.
func NewExecutor(agentRunnerPath string, projectRoot string) *Executor {
	return &Executor{
		AgentRunnerPath: agentRunnerPath,
		ProjectRoot:     projectRoot,
		logger:          logging.WithComponent(slog.Default(), "orchestrator-executor"),
		events:          nil, // Set via SetEventEmitter if needed
	}
}

// SetEventEmitter sets the event emitter for streaming logs
func (e *Executor) SetEventEmitter(emitter EventEmitter) {
	e.events = emitter
}

// SetLogger sets a custom logger for the executor
func (e *Executor) SetLogger(logger *slog.Logger) {
	e.logger = logging.WithComponent(logger, "orchestrator-executor")
}

// ExecuteTask runs the agent-runner for a given task.
func (e *Executor) ExecuteTask(ctx context.Context, task *Task) (*Attempt, error) {
	logger := logging.WithTraceID(e.logger, ctx)
	start := time.Now()

	// Create new attempt
	attempt := &Attempt{
		ID:        uuid.New().String(),
		TaskID:    task.ID,
		Status:    AttemptStatusRunning,
		StartedAt: time.Now(),
	}

	logger.Info("starting task execution",
		slog.String("task_id", task.ID),
		slog.String("task_title", task.Title),
		slog.String("attempt_id", attempt.ID),
	)

	// Update task status to RUNNING (in-memory only, caller handles persistence)
	task.Status = TaskStatusRunning
	now := time.Now()
	task.StartedAt = &now
	logger.Info("task status updated to RUNNING")

	// Generate task YAML for agent-runner
	taskYAML := e.generateTaskYAML(task)
	logger.Debug("generated task YAML", slog.Int("yaml_length", len(taskYAML)))

	// Execute agent-runner
	logger.Info("executing agent-runner", slog.String("binary_path", e.AgentRunnerPath))
	cmd := exec.CommandContext(ctx, e.AgentRunnerPath)
	cmd.Dir = e.ProjectRoot

	// Pass task YAML via stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		logger.Error("failed to create stdin pipe", slog.Any("error", err))
		return e.handleExecutionError(attempt, task, err)
	}

	// Setup stdout/stderr streaming if event emitter is available
	var stdoutPipe, stderrPipe io.ReadCloser
	var outputBuf bytes.Buffer
	if e.events != nil {
		stdoutPipe, err = cmd.StdoutPipe()
		if err != nil {
			logger.Error("failed to create stdout pipe", slog.Any("error", err))
			return e.handleExecutionError(attempt, task, err)
		}
		stderrPipe, err = cmd.StderrPipe()
		if err != nil {
			logger.Error("failed to create stderr pipe", slog.Any("error", err))
			return e.handleExecutionError(attempt, task, err)
		}

		// Stream stdout
		go func() {
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				line := scanner.Text()
				outputBuf.WriteString(line + "\n")

				// Try parsing as structured log/event
				var entry map[string]interface{}
				if err := json.Unmarshal([]byte(line), &entry); err == nil {
					e.handleStructuredLog(task.ID, entry)
				}

				e.events.Emit(EventTaskLog, TaskLogEvent{
					TaskID:    task.ID,
					Stream:    "stdout",
					Line:      line,
					Timestamp: time.Now(),
				})
			}
		}()

		// Stream stderr
		go func() {
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				line := scanner.Text()
				outputBuf.WriteString(line + "\n")
				e.events.Emit(EventTaskLog, TaskLogEvent{
					TaskID:    task.ID,
					Stream:    "stderr",
					Line:      line,
					Timestamp: time.Now(),
				})
			}
		}()
	} else {
		// Fallback: use CombinedOutput if no event emitter
		cmd.Stdout = &outputBuf
		cmd.Stderr = &outputBuf
	}

	go func() {
		defer func() { _ = stdin.Close() }()
		select {
		case <-ctx.Done():
			return // Context canceled, close stdin (via defer) and exit
		default:
			// Write task YAML
			_, _ = stdin.Write([]byte(taskYAML))
		}
	}()

	err = cmd.Start()
	if err != nil {
		logger.Error("failed to start agent-runner", slog.Any("error", err))
		return e.handleExecutionError(attempt, task, err)
	}

	err = cmd.Wait()
	finishedAt := time.Now()
	attempt.FinishedAt = &finishedAt
	output := outputBuf.String()

	if err != nil {
		attempt.Status = AttemptStatusFailed
		attempt.ErrorSummary = fmt.Sprintf("Execution failed: %s\nOutput: %s", err.Error(), string(output))
		task.Status = TaskStatusFailed
		task.DoneAt = &finishedAt
		logger.Error("agent-runner execution failed",
			slog.Any("error", err),
			slog.Int("output_length", len(output)),
			logging.LogDuration(start),
		)
		logger.Debug("agent-runner output", slog.String("output", string(output)))
	} else {
		attempt.Status = AttemptStatusSucceeded
		task.Status = TaskStatusSucceeded
		task.DoneAt = &finishedAt
		logger.Info("agent-runner execution succeeded",
			slog.Int("output_length", len(output)),
			logging.LogDuration(start),
		)
		logger.Debug("agent-runner output", slog.String("output", string(output)))
	}

	// Save updated attempt and task -> REMOVED (Caller responsibility)
	// if err := e.TaskStore.SaveAttempt(attempt); err != nil {
	// 	logger.Error("failed to update attempt", slog.Any("error", err))
	// 	return attempt, fmt.Errorf("failed to update attempt: %w", err)
	// }
	// if err := e.TaskStore.SaveTask(task); err != nil {
	// 	logger.Error("failed to update task", slog.Any("error", err))
	// 	return attempt, fmt.Errorf("failed to update task: %w", err)
	// }

	logger.Info("task execution completed",
		slog.String("final_status", string(attempt.Status)),
		logging.LogDuration(start),
	)
	return attempt, err
}

func (e *Executor) handleExecutionError(attempt *Attempt, task *Task, err error) (*Attempt, error) {
	now := time.Now()
	attempt.FinishedAt = &now
	attempt.Status = AttemptStatusFailed
	attempt.ErrorSummary = err.Error()

	task.Status = TaskStatusFailed
	task.DoneAt = &now

	// _ = e.TaskStore.SaveAttempt(attempt)
	// _ = e.TaskStore.SaveTask(task)

	return attempt, err
}

func (e *Executor) generateTaskYAML(task *Task) string {
	// Construct the prompt text with Description, AcceptanceCriteria, and SuggestedImpl
	promptText := fmt.Sprintf("Execute task: %s", task.Title)
	if task.Description != "" {
		promptText += fmt.Sprintf("\n\nDescription:\n%s", task.Description)
	}
	if len(task.AcceptanceCriteria) > 0 {
		promptText += "\n\nAcceptance Criteria:"
		for _, ac := range task.AcceptanceCriteria {
			promptText += fmt.Sprintf("\n- %s", ac)
		}
	}
	if task.SuggestedImpl != nil {
		promptText += "\n\nSuggested Implementation:"
		if task.SuggestedImpl.Language != "" {
			promptText += fmt.Sprintf("\nLanguage: %s", task.SuggestedImpl.Language)
		}
		if len(task.SuggestedImpl.FilePaths) > 0 {
			promptText += "\nTarget Files:"
			for _, f := range task.SuggestedImpl.FilePaths {
				promptText += fmt.Sprintf("\n- %s", f)
			}
		}
		if len(task.SuggestedImpl.Constraints) > 0 {
			promptText += "\nConstraints:"
			for _, c := range task.SuggestedImpl.Constraints {
				promptText += fmt.Sprintf("\n- %s", c)
			}
		}
	}

	// Simple task YAML for agent-runner
	// Using literal style Block Scalar (|) for prd.text to handle multi-line strings safely.
	// We also populate V2 fields in the YAML.

	promptTextIndented := ""
	for _, line := range strings.Split(promptText, "\n") {
		promptTextIndented += fmt.Sprintf("      %s\n", line)
	}

	// Marshaling SuggestedImpl to YAML manually or via helper would be cleaner,
	// but sticking to string formatting for dependency simplicity as per current pattern.
	suggestedImplYAML := ""
	if task.SuggestedImpl != nil {
		suggestedImplYAML = fmt.Sprintf(`  suggested_impl:
    language: %q
    file_paths: [%s]
    constraints: [%s]`,
			task.SuggestedImpl.Language,
			quoteList(task.SuggestedImpl.FilePaths),
			quoteList(task.SuggestedImpl.Constraints),
		)
	}

	// Dependencies
	dependenciesYAML := fmt.Sprintf("dependencies: [%s]", quoteList(task.Dependencies))

	return fmt.Sprintf(`version: "1"
task:
  id: %s
  title: %q
  repo: "."
  description: %q
  wbs_level: %d
  phase_name: %q
  %s
%s
  prd:
    text: |
%srunner:
  max_loops: 5
  worker:
    cli: "codex"
`, task.ID, task.Title, task.Description, task.WBSLevel, task.PhaseName, dependenciesYAML, suggestedImplYAML, promptTextIndented)
}

func quoteList(items []string) string {
	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = fmt.Sprintf("%q", item)
	}
	return strings.Join(quoted, ", ")
}

func (e *Executor) handleStructuredLog(taskID string, entry map[string]interface{}) {
	eventType, ok := entry["event_type"].(string)
	if !ok {
		return
	}

	timestamp := time.Now()
	if tsStr, ok := entry["time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, tsStr); err == nil {
			timestamp = t
		}
	}

	switch eventType {
	case "meta:thinking":
		detail, _ := entry["detail"].(string)
		e.events.Emit(EventProcessMetaUpdate, ProcessMetaUpdateEvent{
			TaskID:    taskID,
			State:     "THINKING",
			Detail:    detail,
			Timestamp: timestamp,
		})
	case "meta:state_change":
		// Only distinct states, maybe map "state transition" later if needed
	case "container:starting":
		e.events.Emit(EventProcessContainerUpdate, ProcessContainerUpdateEvent{
			TaskID:    taskID,
			Status:    "STARTING",
			Image:     "unknown", // Could add to log if needed
			Timestamp: timestamp,
		})
	case "container:started":
		e.events.Emit(EventProcessContainerUpdate, ProcessContainerUpdateEvent{
			TaskID:      taskID,
			ContainerID: "running", // Don't have ID in log yet, but status is key
			Status:      "RUNNING",
			Timestamp:   timestamp,
		})
	case "worker:running":
		cmd, _ := entry["command"].(string)
		e.events.Emit(EventProcessWorkerUpdate, ProcessWorkerUpdateEvent{
			TaskID:    taskID,
			WorkerID:  "worker-1",
			Status:    "RUNNING",
			Command:   cmd,
			Timestamp: timestamp,
		})
	case "worker:completed":
		exitCode, _ := entry["exit_code"].(float64)
		e.events.Emit(EventProcessWorkerUpdate, ProcessWorkerUpdateEvent{
			TaskID:    taskID,
			WorkerID:  "worker-1",
			Status:    "IDLE", // Or FINISHED
			ExitCode:  int(exitCode),
			Timestamp: timestamp,
		})
	}
}
