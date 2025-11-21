package meta

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Client struct {
	kind   string
	apiKey string
	model  string
	client *http.Client
}

func NewClient(kind, apiKey, model string) *Client {
	if model == "" {
		model = "gpt-4-turbo" // Default
	}
	return &Client{
		kind:   kind,
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 60 * time.Second},
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

func (c *Client) callLLM(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
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

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("OpenAI API error: %s (and failed to read error body: %w)", resp.Status, err)
		}
		return "", fmt.Errorf("OpenAI API error: %s %s", resp.Status, string(body))
	}

	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from LLM")
	}

	return result.Choices[0].Message.Content, nil
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

	systemPrompt := `You are a Meta-agent that plans software development tasks.
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

	systemPrompt := `You are a Meta-agent that orchestrates a coding task.
Decide the next action based on the current context.
Output MUST be a YAML block with type: next_action.
`
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

	systemPrompt := `You are a Meta-agent evaluating task completion.
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
