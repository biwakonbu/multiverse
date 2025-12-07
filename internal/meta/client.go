package meta

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/biwakonbu/agent-runner/internal/logging"
	"gopkg.in/yaml.v3"
)

type Client struct {
	kind         string
	apiKey       string
	model        string
	systemPrompt string
	client       *http.Client
	logger       *slog.Logger
	cliProvider  CLIProvider // CLI プロバイダ（codex-cli 等）
}

func NewClient(kind, apiKey, model, systemPrompt string) *Client {
	if model == "" {
		model = "gpt-5.1-codex-max-high" // Default
	}
	c := &Client{
		kind:         kind,
		apiKey:       apiKey,
		model:        model,
		systemPrompt: systemPrompt,
		client:       &http.Client{Timeout: 60 * time.Second},
		logger:       logging.WithComponent(slog.Default(), "meta-client"),
	}

	// CLI プロバイダの初期化
	if kind == "codex-cli" {
		codexProvider := NewCodexCLIProvider(model, systemPrompt)
		codexProvider.SetLogger(c.logger)
		c.cliProvider = codexProvider
	}

	return c
}

// SetLogger sets a custom logger for the client
func (c *Client) SetLogger(logger *slog.Logger) {
	c.logger = logging.WithComponent(logger, "meta-client")
}

// ... (keep existing structs)

// We need to replace the whole file or use targeted replaces.
// Let's use targeted replaces for PlanTask and NextAction.

// OpenAI Chat Completion Request
type chatRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
}

// isRetryableError determines if an error or status code should trigger a retry
func isRetryableError(err error, resp *http.Response) bool {
	// Check network/timeout errors
	if err != nil {
		// Timeout errors
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return true
		}

		// context cancellation is not retryable
		if err == context.Canceled {
			return false
		}
		// context deadline exceeded is retryable
		if err == context.DeadlineExceeded {
			return true
		}
		// Other errors (like connection refused) may be transient
		return true
	}

	// Check HTTP status codes
	if resp != nil {
		// 5xx errors are retryable
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			return true
		}
		// 429 Too Many Requests (Rate Limit) is retryable
		if resp.StatusCode == 429 {
			return true
		}
	}

	return false
}

func (c *Client) callLLM(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	const maxRetries = 3
	const baseDelay = 1 * time.Second

	logger := logging.WithTraceID(c.logger, ctx)
	start := time.Now()

	reqBody := chatRequest{
		Model: c.model,
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	logger.Info("calling LLM",
		slog.String("model", c.model),
		slog.Int("request_size", len(jsonBody)),
	)
	logger.Debug("LLM request",
		slog.String("system_prompt", systemPrompt),
		slog.String("user_prompt", userPrompt),
	)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Create request fresh for each attempt
		req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			if !isRetryableError(err, nil) {
				return "", err
			}

			// Retryable error, continue to next attempt
			if attempt < maxRetries {
				delay := baseDelay * time.Duration(1<<uint(attempt))
				logger.Warn("LLM request failed, retrying",
					slog.Int("attempt", attempt+1),
					slog.Int("max_retries", maxRetries),
					slog.Float64("delay_seconds", delay.Seconds()),
					slog.Any("error", err),
				)
				select {
				case <-time.After(delay):
					// Continue to next attempt
				case <-ctx.Done():
					return "", ctx.Err()
				}
			}
			continue
		}

		// Close response body on defer, but only if we have a response
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode != 200 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				lastErr = fmt.Errorf("OpenAI API error: %s (and failed to read error body: %w)", resp.Status, err)
			} else {
				lastErr = fmt.Errorf("OpenAI API error: %s %s", resp.Status, string(body))
			}

			if !isRetryableError(nil, resp) {
				// Non-retryable error (4xx, 3xx, etc.), return immediately
				return "", lastErr
			}

			// Retryable error (5xx, 429), continue to next attempt
			if attempt < maxRetries {
				delay := baseDelay * time.Duration(1<<uint(attempt))
				logger.Warn("LLM request failed with retryable status, retrying",
					slog.Int("attempt", attempt+1),
					slog.Int("max_retries", maxRetries),
					slog.Int("status_code", resp.StatusCode),
					slog.Float64("delay_seconds", delay.Seconds()),
				)
				select {
				case <-time.After(delay):
					// Continue to next attempt
				case <-ctx.Done():
					return "", ctx.Err()
				}
			}
			continue
		}

		// Success
		var result chatResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return "", err
		}

		if len(result.Choices) == 0 {
			return "", fmt.Errorf("no choices returned from LLM")
		}

		responseContent := result.Choices[0].Message.Content
		logger.Info("LLM call completed",
			slog.Int("response_size", len(responseContent)),
			logging.LogDuration(start),
		)
		logger.Debug("LLM response", slog.String("content", responseContent))
		return responseContent, nil
	}

	// Max retries exceeded
	if lastErr != nil {
		return "", fmt.Errorf("LLM request failed after %d retries: %w", maxRetries, lastErr)
	}
	return "", fmt.Errorf("LLM request failed after %d retries", maxRetries)
}

// extractYAML extracts YAML content from LLM response, handling markdown code blocks
func extractYAML(response string) string {
	response = strings.TrimSpace(response)

	// Method 1: Try to extract from markdown code block (```yaml ... ```)
	reMarkdown := regexp.MustCompile("(?s)```ya?ml\\s*\\n(.+?)\\n```")
	matches := reMarkdown.FindStringSubmatch(response)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Method 2: Try generic code block extraction (``` ... ```)
	reGeneric := regexp.MustCompile("(?s)```\\s*\\n(.+?)\\n```")
	matches = reGeneric.FindStringSubmatch(response)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Method 3: Strip leading/trailing backticks if present (e.g. `yaml ... ` or ``` ... ``` without newlines)
	// This handles cases where LLM might output inline code or malformed blocks
	if strings.HasPrefix(response, "```") && strings.HasSuffix(response, "```") {
		response = strings.TrimPrefix(response, "```yaml")
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		return strings.TrimSpace(response)
	}

	return response
}

func (c *Client) PlanTask(ctx context.Context, prdText string) (*PlanTaskResponse, error) {
	if c.kind == "mock" {
		return &PlanTaskResponse{
			TaskID: "TASK-MOCK",
			AcceptanceCriteria: []AcceptanceCriterion{
				{ID: "AC-1", Description: "Mock AC 1", Type: "mock", Critical: true},
			},
		}, nil
	}

	// CLI プロバイダを使用する場合
	if c.cliProvider != nil {
		return c.cliProvider.PlanTask(ctx, prdText)
	}

	// HTTP ベースの LLM 呼び出し（後方互換性のため残す）
	systemPrompt := c.systemPrompt
	if systemPrompt == "" {
		systemPrompt = `You are a Meta-agent that plans software development tasks.
Your goal is to read a PRD and break it down into Acceptance Criteria.
Output MUST be a YAML block with the following structure:
type: plan_task
version: 1
payload:
  task_id: "TASK-..."
  acceptance_criteria:
    - id: "AC-1"
      description: "..."
      type: "e2e"
      critical: true
`
	}
	userPrompt := fmt.Sprintf("PRD:\n%s\n\nGenerate the plan.", prdText)

	resp, err := c.callLLM(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	// Extract YAML from response (handles markdown code blocks)
	resp = extractYAML(resp)

	var msg MetaMessage
	if err := yaml.Unmarshal([]byte(resp), &msg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w\nResponse: %s", err, resp)
	}

	// Re-marshal payload to specific struct
	payloadBytes, err := yaml.Marshal(msg.Payload)
	if err != nil {
		return nil, err
	}
	var plan PlanTaskResponse
	if err := yaml.Unmarshal(payloadBytes, &plan); err != nil {
		return nil, err
	}

	return &plan, nil
}

func (c *Client) NextAction(ctx context.Context, taskSummary *TaskSummary) (*NextActionResponse, error) {
	if c.kind == "mock" {
		// Simple mock logic: Run worker once, then complete
		if taskSummary.WorkerRunsCount == 0 {
			return &NextActionResponse{
				Decision: Decision{Action: "run_worker", Reason: "Mock run"},
				WorkerCall: WorkerCall{
					WorkerType: "codex-cli",
					Mode:       "exec",
					Prompt:     "echo 'Hello from Mock Worker'",
				},
			}, nil
		}
		return &NextActionResponse{
			Decision: Decision{Action: "mark_complete", Reason: "Mock complete"},
		}, nil
	}

	// CLI プロバイダを使用する場合
	if c.cliProvider != nil {
		return c.cliProvider.NextAction(ctx, taskSummary)
	}

	// HTTP ベースの LLM 呼び出し（後方互換性のため残す）
	systemPrompt := c.systemPrompt
	if systemPrompt == "" {
		systemPrompt = `You are a Meta-agent that orchestrates a coding task.
Decide the next action based on the current context.
Output MUST be a YAML block with type: next_action.

Schema for 'run_worker' action:
type: next_action
decision:
  action: run_worker
  reason: <string>
worker_call:
  worker_type: codex-cli
  mode: exec
  prompt: <string>
  # Optional fields:
  model: <string>
  flags: [<string>]
  env: {<key>: <value>}
  use_stdin: <bool>

Schema for 'mark_complete' action:
type: next_action
decision:
  action: mark_complete
  reason: <string>
`
	}
	// Serialize context for LLM
	contextSummary := fmt.Sprintf("Task: %s\nState: %s\nACs: %v\nWorkerRuns: %d",
		taskSummary.Title, taskSummary.State, len(taskSummary.AcceptanceCriteria), taskSummary.WorkerRunsCount)

	userPrompt := fmt.Sprintf("Context:\n%s\n\nDecide next action.", contextSummary)

	resp, err := c.callLLM(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	// Extract YAML from response (handles markdown code blocks)
	resp = extractYAML(resp)

	var msg MetaMessage
	if err := yaml.Unmarshal([]byte(resp), &msg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w\nResponse: %s", err, resp)
	}

	payloadBytes, err := yaml.Marshal(msg.Payload)
	if err != nil {
		return nil, err
	}
	var action NextActionResponse
	if err := yaml.Unmarshal(payloadBytes, &action); err != nil {
		return nil, err
	}

	return &action, nil
}

func (c *Client) CompletionAssessment(ctx context.Context, taskSummary *TaskSummary) (*CompletionAssessmentResponse, error) {
	if c.kind == "mock" {
		// Mock implementation: all criteria passed
		criteria := []CriterionResult{}
		for _, ac := range taskSummary.AcceptanceCriteria {
			criteria = append(criteria, CriterionResult{
				ID:      ac.ID,
				Status:  "passed",
				Comment: "Mock assessment: passed",
			})
		}
		return &CompletionAssessmentResponse{
			AllCriteriaSatisfied: true,
			Summary:              "Mock: All criteria satisfied",
			ByCriterion:          criteria,
		}, nil
	}

	// CLI プロバイダを使用する場合
	if c.cliProvider != nil {
		return c.cliProvider.CompletionAssessment(ctx, taskSummary)
	}

	// HTTP ベースの LLM 呼び出し（後方互換性のため残す）
	systemPrompt := c.systemPrompt
	if systemPrompt == "" {
		systemPrompt = `You are a Meta-agent evaluating task completion.
Review the Acceptance Criteria and Worker execution results.
Output MUST be a YAML block with type: completion_assessment.

Example format:
type: completion_assessment
version: 1
payload:
  all_criteria_satisfied: true
  summary: "All acceptance criteria met"
  by_criterion:
    - id: "AC-1"
      status: "passed"
      comment: "Feature X successfully implemented"
`
	}

	// Format acceptance criteria for LLM
	acText := ""
	for _, ac := range taskSummary.AcceptanceCriteria {
		acText += fmt.Sprintf("- %s: %s\n", ac.ID, ac.Description)
	}

	// Format worker runs for LLM
	workerText := ""
	for _, run := range taskSummary.WorkerRuns {
		workerText += fmt.Sprintf("- Run %s: exit_code=%d, summary=%s\n", run.ID, run.ExitCode, run.Summary)
	}

	userPrompt := fmt.Sprintf(`Task: %s
State: %s

Acceptance Criteria:
%s

Worker Execution Results:
%s

Evaluate whether all acceptance criteria are satisfied.`,
		taskSummary.Title, taskSummary.State, acText, workerText)

	resp, err := c.callLLM(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	// Extract YAML from response (handles markdown code blocks)
	resp = extractYAML(resp)

	var msg MetaMessage
	if err := yaml.Unmarshal([]byte(resp), &msg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w\nResponse: %s", err, resp)
	}

	// Re-marshal payload to specific struct
	payloadBytes, err := yaml.Marshal(msg.Payload)
	if err != nil {
		return nil, err
	}
	var assessment CompletionAssessmentResponse
	if err := yaml.Unmarshal(payloadBytes, &assessment); err != nil {
		return nil, err
	}

	return &assessment, nil
}

// Decompose はユーザー入力からタスクを分解する（v2.0 チャット駆動）
func (c *Client) Decompose(ctx context.Context, req *DecomposeRequest) (*DecomposeResponse, error) {
	logger := logging.WithTraceID(c.logger, ctx)

	if c.kind == "mock" {
		logger.Debug("using mock decompose response")
		return &DecomposeResponse{
			Understanding: "Mock: ユーザーの要求を理解しました",
			Phases: []DecomposedPhase{
				{
					Name:      "概念設計",
					Milestone: "M1-Mock-Design",
					Tasks: []DecomposedTask{
						{
							ID:                 "temp-001",
							Title:              "Mock概念設計タスク",
							Description:        "モック用の概念設計タスクです",
							AcceptanceCriteria: []string{"設計ドキュメントが作成されている"},
							Dependencies:       []string{},
							WBSLevel:           1,
							EstimatedEffort:    "small",
						},
					},
				},
				{
					Name:      "実装",
					Milestone: "M2-Mock-Impl",
					Tasks: []DecomposedTask{
						{
							ID:                 "temp-002",
							Title:              "Mock実装タスク",
							Description:        "モック用の実装タスクです",
							AcceptanceCriteria: []string{"機能が実装されている", "テストが通過している"},
							Dependencies:       []string{"temp-001"},
							WBSLevel:           3,
							EstimatedEffort:    "medium",
						},
					},
				},
			},
			PotentialConflicts: []PotentialConflict{},
		}, nil
	}

	// CLI プロバイダを使用する場合
	if c.cliProvider != nil {
		return c.cliProvider.Decompose(ctx, req)
	}

	// HTTP ベースの LLM 呼び出し（後方互換性のため残す）
	systemPrompt := decomposeSystemPrompt
	userPrompt := buildDecomposeUserPrompt(req)

	logger.Info("calling LLM for decompose",
		slog.String("user_input_length", fmt.Sprintf("%d", len(req.UserInput))),
		slog.Int("existing_tasks", len(req.Context.ExistingTasks)),
	)

	resp, err := c.callLLM(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// YAML を抽出してパース
	resp = extractYAML(resp)

	var msg MetaMessage
	if err := yaml.Unmarshal([]byte(resp), &msg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w\nResponse: %s", err, resp)
	}

	payloadBytes, err := yaml.Marshal(msg.Payload)
	if err != nil {
		return nil, err
	}

	var decompose DecomposeResponse
	if err := yaml.Unmarshal(payloadBytes, &decompose); err != nil {
		return nil, fmt.Errorf("failed to parse decompose response: %w", err)
	}

	logger.Info("decompose completed",
		slog.Int("phases", len(decompose.Phases)),
		slog.Int("potential_conflicts", len(decompose.PotentialConflicts)),
	)

	return &decompose, nil
}

// decomposeSystemPrompt はタスク分解用のシステムプロンプト
const decomposeSystemPrompt = `You are a Meta-agent that decomposes user requests into structured development tasks.

Your goal is to:
1. Understand the user's intent from their message
2. Break down the request into phases: 概念設計 (Conceptual Design) → 実装設計 (Implementation Design) → 実装 (Implementation)
3. Create detailed tasks with clear acceptance criteria
4. Identify dependencies between tasks
5. Flag potential file conflicts

Output MUST be a YAML block with the following structure:
type: decompose
version: 1
payload:
  understanding: "ユーザーの要求を理解した内容..."
  phases:
    - name: "概念設計"
      milestone: "M1-Feature-Design"
      tasks:
        - id: "temp-001"
          title: "タスクタイトル"
          description: "詳細な説明"
          acceptance_criteria:
            - "達成条件1"
            - "達成条件2"
          dependencies: []
          wbs_level: 1
          estimated_effort: "small"
    - name: "実装設計"
      milestone: "M1-Feature-Design"
      tasks:
        - id: "temp-002"
          title: "..."
          description: "..."
          acceptance_criteria: [...]
          dependencies: ["temp-001"]
          wbs_level: 2
          estimated_effort: "medium"
    - name: "実装"
      milestone: "M2-Feature-Impl"
      tasks:
        - id: "temp-003"
          title: "..."
          description: "..."
          acceptance_criteria: [...]
          dependencies: ["temp-002"]
          wbs_level: 3
          estimated_effort: "large"
  potential_conflicts:
    - file: "src/example.ts"
      tasks: ["temp-003"]
      warning: "既存ファイルを変更する可能性があります"

Guidelines:
- WBS levels: 1=概念設計, 2=実装設計, 3=実装
- Estimated effort: small (< 1 hour), medium (1-4 hours), large (> 4 hours)
- Task IDs must start with "temp-" (they will be replaced with permanent IDs)
- Dependencies can reference other temp IDs from the same batch
- Be specific about acceptance criteria - they should be verifiable
- Consider existing tasks to avoid duplication
`

// buildDecomposeUserPrompt はユーザープロンプトを構築する
func buildDecomposeUserPrompt(req *DecomposeRequest) string {
	var sb strings.Builder

	sb.WriteString("## User Request\n")
	sb.WriteString(req.UserInput)
	sb.WriteString("\n\n")

	sb.WriteString("## Context\n")
	sb.WriteString(fmt.Sprintf("Workspace: %s\n\n", req.Context.WorkspacePath))

	if len(req.Context.ExistingTasks) > 0 {
		sb.WriteString("### Existing Tasks\n")
		for _, task := range req.Context.ExistingTasks {
			deps := ""
			if len(task.Dependencies) > 0 {
				deps = fmt.Sprintf(" (depends: %s)", strings.Join(task.Dependencies, ", "))
			}
			sb.WriteString(fmt.Sprintf("- [%s] %s: %s%s\n", task.Status, task.ID, task.Title, deps))
		}
		sb.WriteString("\n")
	}

	if len(req.Context.ConversationHistory) > 0 {
		sb.WriteString("### Conversation History\n")
		for _, msg := range req.Context.ConversationHistory {
			sb.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Please decompose this request into structured tasks.")
	return sb.String()
}

// NewMockClient creates a mock Meta client (kind="mock")
func NewMockClient() *Client {
	return &Client{
		kind:   "mock",
		apiKey: "",
		model:  "mock",
		client: &http.Client{Timeout: 60 * time.Second},
		logger: logging.WithComponent(slog.Default(), "meta-client-mock"),
	}
}
