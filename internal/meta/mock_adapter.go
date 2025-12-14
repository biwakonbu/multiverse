package meta

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/biwakonbu/agent-runner/internal/logging"
)

// NewMockClient creates a Client that simulates the old mock behavior using a custom Transport.
// This allows removing "if kind == mock" checks from the main logic while keeping tests passing.
func NewMockClient() *Client {
	// Use OpenAIProvider with a mock Transport to simulate responses
	provider := &OpenAIProvider{
		apiKey:       "MOCK_KEY", // Dummy key to pass validation in Provider
		model:        "mock",
		systemPrompt: "",
		client: &http.Client{
			Timeout:   60 * time.Second,
			Transport: &shimMockRoundTripper{},
		},
		logger: logging.WithComponent(slog.Default(), "meta-openai-mock"),
	}

	return &Client{
		kind:         "mock",
		apiKey:       "MOCK_KEY",
		model:        "mock",
		systemPrompt: "",
		provider:     provider,
		logger:       logging.WithComponent(slog.Default(), "meta-client-mock"),
	}
}

// shimMockRoundTripper intercepts HTTP requests and returns hardcoded responses
// mirroring the old "mock mode" logic.
// QH-005 (PRD 13.3 #4): Uses structure-based matching with string fallback
type shimMockRoundTripper struct{}

// openAIRequest represents the structure of OpenAI chat completion request
type openAIRequest struct {
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	Model string `json:"model"`
}

func (m *shimMockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Read body to determine request type
	bodyBytes, _ := io.ReadAll(req.Body)
	_ = req.Body.Close()
	// Restore body for any downstream readers (not needed here but good practice)
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// QH-005: Try structure-based matching first
	var oaiReq openAIRequest
	_ = json.Unmarshal(bodyBytes, &oaiReq) // ignore error, fallback to string matching

	bodyStr := string(bodyBytes)

	// Extract system prompt from messages (first message with role="system")
	systemPrompt := ""
	userPrompt := ""
	for _, msg := range oaiReq.Messages {
		if msg.Role == "system" && systemPrompt == "" {
			systemPrompt = msg.Content
		}
		if msg.Role == "user" && userPrompt == "" {
			userPrompt = msg.Content
		}
	}

	// Determine response based on content
	var content string

	// QH-005: Structure-based detection with fallback to string matching
	switch {
	case strings.Contains(systemPrompt, "generates a task plan") || strings.Contains(bodyStr, "Generate the plan"):
		// PlanTask
		content = `
type: plan_task
version: 1
payload:
  task_id: "TASK-MOCK"
  acceptance_criteria:
    - id: "AC-1"
      description: "Mock AC 1"
      type: "mock"
      critical: true
`
	case strings.Contains(systemPrompt, "decides what to do next") || strings.Contains(bodyStr, "Decide next action"):
		// NextAction
		// Check context to see if we should run worker or complete
		if strings.Contains(bodyStr, "WorkerRuns: 0") || strings.Contains(userPrompt, "WorkerRuns: 0") {
			content = `
type: next_action
version: 1
payload:
  decision:
    action: "run_worker"
    reason: "Mock run"
  worker_call:
    worker_type: "codex-cli"
    mode: "exec"
    prompt: "echo 'Hello from Mock Worker'"
`
		} else {
			content = `
type: next_action
version: 1
payload:
  decision:
    action: "mark_complete"
    reason: "Mock complete"
`
		}
	case strings.Contains(systemPrompt, "evaluates task completion") || strings.Contains(bodyStr, "Evaluate whether all acceptance criteria are satisfied"):
		// CompletionAssessment
		content = `
type: completion_assessment
version: 1
payload:
  all_criteria_satisfied: true
  summary: "Mock: All criteria satisfied"
  by_criterion:
    - id: "AC-1"
      status: "passed"
      comment: "Mock assessment: passed"
    - id: "AC-2"
      status: "passed"
      comment: "Mock assessment: passed"
`
	case strings.Contains(systemPrompt, "maintains and edits a development plan") || strings.Contains(bodyStr, "\"type\": \"plan_patch\"") || strings.Contains(bodyStr, "plan_patch operations"):
		// PlanPatch
		content = `{
  "type": "plan_patch",
  "version": 1,
  "payload": {
    "understanding": "Mock: ユーザーの要求を理解しました",
    "operations": [
      {
        "op": "create",
        "temp_id": "temp-001",
        "title": "Mock概念設計タスク",
        "description": "モック用の概念設計タスクです",
        "acceptance_criteria": ["設計ドキュメントが作成されている"],
        "dependencies": [],
        "wbs_level": 1,
        "phase_name": "概念設計",
        "milestone": "M1-Mock-Design"
      },
      {
        "op": "create",
        "temp_id": "temp-002",
        "title": "Mock実装タスク",
        "description": "モック用の実装タスクです",
        "acceptance_criteria": ["機能が実装されている", "テストが通過している"],
        "dependencies": ["temp-001"],
        "wbs_level": 3,
        "phase_name": "実装",
        "milestone": "M2-Mock-Impl",
        "suggested_impl": {
          "language": "go",
          "file_paths": ["internal/mock/mock.go"],
          "constraints": ["Keep backward compatibility"]
        }
      }
    ],
    "potential_conflicts": []
  }
}`
	case strings.Contains(systemPrompt, "decomposes user requests") || strings.Contains(bodyStr, "decompose this request"):
		// Decompose
		content = `{
  "type": "decompose",
  "version": 1,
  "payload": {
    "understanding": "Mock: ユーザーの要求を理解しました",
    "phases": [
      {
        "name": "概念設計",
        "milestone": "M1-Mock-Design",
        "tasks": [
          {
            "id": "temp-001",
            "title": "Mock概念設計タスク",
            "description": "モック用の概念設計タスクです",
            "acceptance_criteria": ["設計ドキュメントが作成されている"],
            "dependencies": [],
            "wbs_level": 1,
            "estimated_effort": "small"
          }
        ]
      },
      {
        "name": "実装",
        "milestone": "M2-Mock-Impl",
        "tasks": [
          {
            "id": "temp-002",
            "title": "Mock実装タスク",
            "description": "モック用の実装タスクです",
            "acceptance_criteria": ["機能が実装されている", "テストが通過している"],
            "dependencies": ["temp-001"],
            "wbs_level": 3,
            "estimated_effort": "medium"
          }
        ]
      }
    ],
    "potential_conflicts": []
  }
}`
	default:
		// Default fallback
		content = "Mock response"
	}

	// Wrap in OpenAI format
	jsonResp := `{"choices":[{"message":{"role":"assistant","content":` + quoteString(content) + `}}]}`

	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(jsonResp)),
		Header:     make(http.Header),
	}, nil
}

func quoteString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
