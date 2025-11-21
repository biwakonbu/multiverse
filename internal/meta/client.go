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

	"gopkg.in/yaml.v3"
)

type Client struct {
	kind         string
	apiKey       string
	model        string
	systemPrompt string
	client       *http.Client
}

func NewClient(kind, apiKey, model, systemPrompt string) *Client {
	if model == "" {
		model = "gpt-4-turbo" // Default
	}
	return &Client{
		kind:         kind,
		apiKey:       apiKey,
		model:        model,
		systemPrompt: systemPrompt,
		client:       &http.Client{Timeout: 60 * time.Second},
	}
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
				slog.Warn("LLM request failed, retrying",
					"attempt", attempt+1,
					"max_retries", maxRetries,
					"delay_seconds", delay.Seconds(),
					"error", err.Error())
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
				slog.Warn("LLM request failed with retryable status, retrying",
					"attempt", attempt+1,
					"max_retries", maxRetries,
					"status_code", resp.StatusCode,
					"delay_seconds", delay.Seconds())
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

		return result.Choices[0].Message.Content, nil
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

	// Method 3: Strip leading/trailing backticks if present
	response = strings.TrimPrefix(response, "```yaml")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

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

	systemPrompt := c.systemPrompt
	if systemPrompt == "" {
		systemPrompt = `You are a Meta-agent that orchestrates a coding task.
Decide the next action based on the current context.
Output MUST be a YAML block with type: next_action.
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

// NewMockClient creates a mock Meta client (kind="mock")
func NewMockClient() *Client {
	return &Client{
		kind:   "mock",
		apiKey: "",
		model:  "mock",
		client: &http.Client{Timeout: 60 * time.Second},
	}
}
