package meta

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// Helper to create a client that acts like the old mock for tests.
// Since we removed mock logic from Client methods, we need to mock the HTTP transport
// or use a different way to test "mock" behavior if the test relies on it.
// However, the existing tests call NewMockClient expecting internal mock logic.
// We must recreate that logic or update tests to invoke the new CliProvider logic with a mock provider?
// Or we can mock the HTTP client in the struct.

// Since the user said "remove mock logic from production code",
// the `PlanTask` etc methods NO LONGER have `if kind == "mock"`.
// So checking `kind == "mock"` in tests won't work unless we mock the network.

// Ideally we should use httptest.Server or a Replaceable Transport.

// TestExtractYAML tests the YAML extraction function
func TestExtractYAML_PlainYAML(t *testing.T) {
	input := `type: plan_task
version: 1
payload:
  task_id: "TASK-123"`

	result := extractYAML(input)
	if result != input {
		t.Errorf("Plain YAML not preserved: got %q", result)
	}
}

func TestExtractYAML_MarkdownYAML(t *testing.T) {
	input := "```yaml\ntype: plan_task\nversion: 1\npayload:\n  task_id: \"TASK-123\"\n```"

	expected := `type: plan_task
version: 1
payload:
  task_id: "TASK-123"`

	result := extractYAML(input)
	if result != expected {
		t.Errorf("Markdown YAML extraction failed: got %q, want %q", result, expected)
	}
}

func TestExtractYAML_MarkdownGenericBlock(t *testing.T) {
	input := "```\ntype: plan_task\nversion: 1\n```"

	expected := `type: plan_task
version: 1`

	result := extractYAML(input)
	if result != expected {
		t.Errorf("Generic code block extraction failed: got %q, want %q", result, expected)
	}
}

func TestExtractYAML_LeadingTrailingBackticks(t *testing.T) {
	input := "```yaml\ntype: plan_task```"

	result := extractYAML(input)
	if result == input {
		t.Errorf("Should remove backticks, got %q", result)
	}
}

func TestExtractYAML_WithWhitespace(t *testing.T) {
	input := "\n\n```yaml\ntype: plan_task\nversion: 1\n```\n\n"

	result := extractYAML(input)
	expected := `type: plan_task
version: 1`

	if result != expected {
		t.Errorf("Whitespace handling failed: got %q, want %q", result, expected)
	}
}

// TestClient_PlanTask_Success tests successful PlanTask with mock mode
func TestClient_PlanTask_Success(t *testing.T) {
	// Use mock mode to avoid HTTP calls
	client := NewMockClient()
	result, err := client.PlanTask(context.Background(), "test prd")

	if err != nil {
		t.Fatalf("PlanTask failed: %v", err)
	}

	if result.TaskID != "TASK-MOCK" {
		t.Errorf("TaskID = %q, want TASK-MOCK", result.TaskID)
	}

	if len(result.AcceptanceCriteria) == 0 {
		t.Errorf("AcceptanceCriteria is empty")
	}

	if result.AcceptanceCriteria[0].ID != "AC-1" {
		t.Errorf("First AC ID = %q, want AC-1", result.AcceptanceCriteria[0].ID)
	}
}

// TestClient_NextAction_Success tests successful NextAction
func TestClient_NextAction_Success(t *testing.T) {
	client := NewMockClient()
	summary := &TaskSummary{
		Title:              "Test Task",
		State:              "RUNNING",
		AcceptanceCriteria: []AcceptanceCriterion{},
		WorkerRunsCount:    0,
	}

	result, err := client.NextAction(context.Background(), summary)

	if err != nil {
		t.Fatalf("NextAction failed: %v", err)
	}

	if result.Decision.Action != "run_worker" {
		t.Errorf("First action should be run_worker, got %q", result.Decision.Action)
	}

	// Call again to test mark_complete
	summary.WorkerRunsCount = 1
	result, err = client.NextAction(context.Background(), summary)

	if err != nil {
		t.Fatalf("Second NextAction failed: %v", err)
	}

	if result.Decision.Action != "mark_complete" {
		t.Errorf("Second action should be mark_complete, got %q", result.Decision.Action)
	}
}

// TestClient_MockMode tests that mock mode doesn't require API key
func TestClient_MockMode_NoAPIKey(t *testing.T) {
	client := NewMockClient()

	if client.kind != "mock" {
		t.Errorf("kind = %q, want mock", client.kind)
	}

	if client.apiKey != "" {
		t.Errorf("mock client should not have API key")
	}
}

// TestClient_NewMockClient tests the NewMockClient factory
func TestClient_NewMockClient(t *testing.T) {
	client := NewMockClient()

	if client == nil {
		t.Fatalf("NewMockClient returned nil")
	}

	if client.kind != "mock" {
		t.Errorf("kind = %q, want mock", client.kind)
	}

	if client.model != "mock" {
		t.Errorf("model = %q, want mock", client.model)
	}

	if client.client == nil {
		t.Errorf("http.Client is nil")
	}
}

// TestClient_PlanTask_MarkdownCodeBlock tests YAML extraction from markdown
func TestClient_PlanTask_MarkdownCodeBlock(t *testing.T) {
	client := NewMockClient()

	// This test verifies that the extractYAML logic is applied.
	// We use mock mode to avoid actual API calls.
	result, err := client.PlanTask(context.Background(), "test prd")

	if err != nil {
		t.Fatalf("PlanTask failed: %v", err)
	}

	if result == nil {
		t.Fatalf("PlanTask returned nil result")
	}

	if result.TaskID == "" {
		t.Errorf("TaskID is empty")
	}
}

// TestClient_NextAction_MultipleRuns tests NextAction behavior with multiple runs
func TestClient_NextAction_MultipleRuns(t *testing.T) {
	client := NewMockClient()
	summary := &TaskSummary{
		Title:              "Test Task",
		State:              "RUNNING",
		AcceptanceCriteria: []AcceptanceCriterion{},
		WorkerRunsCount:    0,
	}

	// First call should return run_worker
	action1, _ := client.NextAction(context.Background(), summary)
	if action1.Decision.Action != "run_worker" {
		t.Errorf("First action should be run_worker")
	}

	// Second call should return mark_complete
	summary.WorkerRunsCount = 1
	action2, _ := client.NextAction(context.Background(), summary)
	if action2.Decision.Action != "mark_complete" {
		t.Errorf("Second action should be mark_complete")
	}

	// Further calls should still return mark_complete
	summary.WorkerRunsCount = 5
	action3, _ := client.NextAction(context.Background(), summary)
	if action3.Decision.Action != "mark_complete" {
		t.Errorf("Subsequent actions should be mark_complete")
	}
}

// TestExtractYAML_WithLeadingWhitespace tests YAML extraction with leading whitespace
func TestExtractYAML_WithLeadingWhitespace(t *testing.T) {
	input := "   ```yaml\ntype: plan_task\nversion: 1\n```   "

	result := extractYAML(input)
	expected := `type: plan_task
version: 1`

	if result != expected {
		t.Errorf("Whitespace handling failed: got %q, want %q", result, expected)
	}
}

// TestExtractYAML_OnlyYAML tests extraction when there's only YAML content
func TestExtractYAML_OnlyYAML(t *testing.T) {
	input := `type: test
data:
  key: value`

	result := extractYAML(input)
	if result != input {
		t.Errorf("Plain YAML should be preserved: got %q", result)
	}
}

// TestExtractYAML_WithMarkdownComments tests YAML in code block with surrounding text
func TestExtractYAML_WithMarkdownComments(t *testing.T) {
	input := "Here's the response:\n\n```yaml\ntype: plan_task\nversion: 1\npayload:\n  task_id: \"TASK-100\"\n```\n\nDone."

	result := extractYAML(input)
	expected := "type: plan_task\nversion: 1\npayload:\n  task_id: \"TASK-100\""

	if result != expected {
		t.Errorf("Markdown with surrounding text failed: got %q, want %q", result, expected)
	}
}

// TestClient_NewMockClient_HasHTTPClient tests that mock client has HTTP client
func TestClient_NewMockClient_HasHTTPClient(t *testing.T) {
	client := NewMockClient()
	if client.client == nil {
		t.Errorf("Mock client should have HTTP client")
	}
}

// TestClient_NewMockClient_HasModel tests that mock client has model set
func TestClient_NewMockClient_HasModel(t *testing.T) {
	client := NewMockClient()
	if client.model != "mock" {
		t.Errorf("Mock client model should be 'mock', got %q", client.model)
	}
}

// TestClient_NextAction_NoWorkerRuns tests NextAction behavior with zero worker runs
func TestClient_NextAction_NoWorkerRuns(t *testing.T) {
	client := NewMockClient()
	summary := &TaskSummary{
		Title:           "Test Task",
		State:           "RUNNING",
		WorkerRunsCount: 0,
	}

	result, err := client.NextAction(context.Background(), summary)
	if err != nil {
		t.Fatalf("NextAction failed: %v", err)
	}

	if result.Decision.Action != "run_worker" {
		t.Errorf("With 0 worker runs, action should be run_worker, got %q", result.Decision.Action)
	}
}

// TestClient_PlanTask_MockResponse validates the mock response structure
func TestClient_PlanTask_MockResponse(t *testing.T) {
	client := NewMockClient()
	result, _ := client.PlanTask(context.Background(), "Any PRD")

	if result == nil {
		t.Fatalf("PlanTask should return non-nil result")
	}

	if result.TaskID == "" {
		t.Errorf("TaskID should not be empty")
	}

	if len(result.AcceptanceCriteria) == 0 {
		t.Errorf("AcceptanceCriteria should not be empty in mock response")
	}

	// Validate first AC
	ac := result.AcceptanceCriteria[0]
	if ac.ID == "" {
		t.Errorf("AC ID should not be empty")
	}
}

// TestClient_Kind tests the client kind setter
func TestClient_Kind(t *testing.T) {
	client := &Client{kind: "mock"}
	if client.kind != "mock" {
		t.Errorf("Client kind should be 'mock', got %q", client.kind)
	}

	client2 := &Client{kind: "openai-chat"}
	if client2.kind != "openai-chat" {
		t.Errorf("Client kind should be 'openai-chat', got %q", client2.kind)
	}
}

// TestExtractYAML_MultipleCodeBlocks tests extraction with multiple code blocks (takes first)
func TestExtractYAML_MultipleCodeBlocks(t *testing.T) {
	input := "First:\n```yaml\ntype: plan_task\nversion: 1\n```\n\nSecond:\n```yaml\ntype: other\n```"

	result := extractYAML(input)
	// Should extract the first one
	expected := "type: plan_task\nversion: 1"
	if result != expected {
		t.Errorf("Should extract first code block, got %q, want %q", result, expected)
	}
}

// TestClient_APIKeyHandling tests that API key is stored correctly
func TestClient_APIKeyHandling(t *testing.T) {
	apiKey := "test-api-key-12345"
	client := NewClient("openai-chat", apiKey, "", "")

	// We can't access private field, but we can verify the client was created
	if client == nil {
		t.Fatalf("Client creation failed")
	}

	// Verify other public fields
	// デフォルトモデルは gpt-5.2（Meta-agent 用）
	if client.model != "gpt-5.2" {
		t.Errorf("expected default model gpt-5.2, got %s", client.model)
	}
}

// mockRoundTripper allows controlling HTTP responses for testing
type mockRoundTripper struct {
	mu                sync.Mutex
	responses         []http.Response
	errors            []error
	callCount         int
	timeBetweenCalls  []time.Time
	returnedResponses int
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.timeBetweenCalls = append(m.timeBetweenCalls, time.Now())
	defer func() {
		m.callCount++
	}()

	if m.returnedResponses < len(m.errors) && m.errors[m.returnedResponses] != nil {
		m.returnedResponses++
		return nil, m.errors[m.returnedResponses-1]
	}

	if m.returnedResponses < len(m.responses) {
		resp := m.responses[m.returnedResponses]
		m.returnedResponses++
		return &resp, nil
	}

	// Default successful response
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(`{"choices":[{"message":{"role":"assistant","content":"test response"}}]}`)),
		Header:     make(http.Header),
	}, nil
}

// mockRoundTripperFunc is a function-based RoundTripper for simple mock implementations
type mockRoundTripperFunc func(*http.Request) (*http.Response, error)

func (f mockRoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// TestIsRetryableError tests the isRetryableError helper function
func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		resp        *http.Response
		shouldRetry bool
	}{
		{
			name:        "No error, 200 response",
			err:         nil,
			resp:        &http.Response{StatusCode: 200},
			shouldRetry: false,
		},
		{
			name:        "No error, 4xx response",
			err:         nil,
			resp:        &http.Response{StatusCode: 400},
			shouldRetry: false,
		},
		{
			name:        "No error, 500 response",
			err:         nil,
			resp:        &http.Response{StatusCode: 500},
			shouldRetry: true,
		},
		{
			name:        "No error, 503 response",
			err:         nil,
			resp:        &http.Response{StatusCode: 503},
			shouldRetry: true,
		},
		{
			name:        "No error, 429 rate limit response",
			err:         nil,
			resp:        &http.Response{StatusCode: 429},
			shouldRetry: true,
		},
		{
			name:        "Timeout error",
			err:         &timeoutError{},
			resp:        nil,
			shouldRetry: true,
		},
		{
			name:        "Context canceled",
			err:         context.Canceled,
			resp:        nil,
			shouldRetry: false,
		},
		{
			name:        "Context deadline exceeded",
			err:         context.DeadlineExceeded,
			resp:        nil,
			shouldRetry: true,
		},
		{
			name:        "Generic error",
			err:         io.EOF,
			resp:        nil,
			shouldRetry: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isRetryableError(test.err, test.resp)
			if result != test.shouldRetry {
				t.Errorf("Expected shouldRetry=%v, got %v", test.shouldRetry, result)
			}
		})
	}
}

// timeoutError implements net.Error interface for testing
type timeoutError struct{}

func (t *timeoutError) Error() string   { return "timeout" }
func (t *timeoutError) Timeout() bool   { return true }
func (t *timeoutError) Temporary() bool { return false }

// TestCallLLM_SuccessFirstAttempt tests that successful first attempt returns immediately
func TestCallLLM_SuccessFirstAttempt(t *testing.T) {
	client := &Client{
		kind:  "openai-chat",
		model: "gpt-4-turbo",
		client: &http.Client{
			Transport: &mockRoundTripper{
				responses: []http.Response{
					{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewBufferString(`{"choices":[{"message":{"role":"assistant","content":"success"}}]}`)),
						Header:     make(http.Header),
					},
				},
			},
		},
	}

	ctx := context.Background()
	result, err := client.callLLM(ctx, "system", "user")

	if err != nil {
		t.Fatalf("callLLM failed: %v", err)
	}

	if result != "success" {
		t.Errorf("Expected 'success', got %q", result)
	}
}

// TestCallLLM_RetryOn5xx tests retry behavior on 5xx errors
func TestCallLLM_RetryOn5xx(t *testing.T) {
	client := &Client{
		kind:  "openai-chat",
		model: "gpt-4-turbo",
		client: &http.Client{
			Transport: &mockRoundTripper{
				responses: []http.Response{
					{
						StatusCode: 503,
						Body:       io.NopCloser(bytes.NewBufferString("Service Unavailable")),
						Header:     make(http.Header),
					},
					{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewBufferString(`{"choices":[{"message":{"role":"assistant","content":"recovered"}}]}`)),
						Header:     make(http.Header),
					},
				},
			},
		},
	}

	ctx := context.Background()
	result, err := client.callLLM(ctx, "system", "user")

	if err != nil {
		t.Fatalf("callLLM failed: %v", err)
	}

	if result != "recovered" {
		t.Errorf("Expected 'recovered', got %q", result)
	}
}

// TestCallLLM_RetryOnRateLimit tests retry behavior on 429 (rate limit) errors
func TestCallLLM_RetryOnRateLimit(t *testing.T) {
	client := &Client{
		kind:  "openai-chat",
		model: "gpt-4-turbo",
		client: &http.Client{
			Transport: &mockRoundTripper{
				responses: []http.Response{
					{
						StatusCode: 429,
						Body:       io.NopCloser(bytes.NewBufferString("Too Many Requests")),
						Header:     make(http.Header),
					},
					{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewBufferString(`{"choices":[{"message":{"role":"assistant","content":"rate limited"}}]}`)),
						Header:     make(http.Header),
					},
				},
			},
		},
	}

	ctx := context.Background()
	result, err := client.callLLM(ctx, "system", "user")

	if err != nil {
		t.Fatalf("callLLM failed: %v", err)
	}

	if result != "rate limited" {
		t.Errorf("Expected 'rate limited', got %q", result)
	}
}

// TestCallLLM_NoRetryOn4xx tests that 4xx errors are not retried
func TestCallLLM_NoRetryOn4xx(t *testing.T) {
	client := &Client{
		kind:  "openai-chat",
		model: "gpt-4-turbo",
		client: &http.Client{
			Transport: &mockRoundTripper{
				responses: []http.Response{
					{
						StatusCode: 401,
						Status:     "401 Unauthorized",
						Body:       io.NopCloser(bytes.NewBufferString("Unauthorized")),
						Header:     make(http.Header),
					},
				},
			},
		},
	}

	ctx := context.Background()
	_, err := client.callLLM(ctx, "system", "user")

	if err == nil {
		t.Fatal("callLLM should have failed with 401 error")
	}

	if !strings.Contains(err.Error(), "Unauthorized") {
		t.Errorf("Expected Unauthorized error, got: %v", err)
	}
}

// TestCallLLM_MaxRetriesExceeded tests failure after max retries
func TestCallLLM_MaxRetriesExceeded(t *testing.T) {
	client := &Client{
		kind:  "openai-chat",
		model: "gpt-4-turbo",
		client: &http.Client{
			Transport: &mockRoundTripper{
				responses: []http.Response{
					{
						StatusCode: 500,
						Body:       io.NopCloser(bytes.NewBufferString("Error 1")),
						Header:     make(http.Header),
					},
					{
						StatusCode: 500,
						Body:       io.NopCloser(bytes.NewBufferString("Error 2")),
						Header:     make(http.Header),
					},
					{
						StatusCode: 500,
						Body:       io.NopCloser(bytes.NewBufferString("Error 3")),
						Header:     make(http.Header),
					},
					{
						StatusCode: 500,
						Body:       io.NopCloser(bytes.NewBufferString("Error 4")),
						Header:     make(http.Header),
					},
				},
			},
		},
	}

	ctx := context.Background()
	_, err := client.callLLM(ctx, "system", "user")

	if err == nil {
		t.Fatal("callLLM should have failed after max retries")
	}

	if !strings.Contains(err.Error(), "failed after 3 retries") {
		t.Errorf("Expected 'failed after 3 retries', got: %v", err)
	}
}

// TestCallLLM_RetryOnTimeout tests retry behavior on timeout errors
func TestCallLLM_RetryOnTimeout(t *testing.T) {
	callCount := 0

	customTransport := mockRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if callCount == 0 {
			callCount++
			return nil, &timeoutError{}
		}
		// Second attempt succeeds
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(`{"choices":[{"message":{"role":"assistant","content":"timeout recovered"}}]}`)),
			Header:     make(http.Header),
		}, nil
	})

	client := &Client{
		kind:  "openai-chat",
		model: "gpt-4-turbo",
		client: &http.Client{
			Transport: customTransport,
		},
	}

	ctx := context.Background()
	result, err := client.callLLM(ctx, "system", "user")

	if err != nil {
		t.Fatalf("callLLM failed: %v", err)
	}

	if result != "timeout recovered" {
		t.Errorf("Expected 'timeout recovered', got %q", result)
	}
}

// TestCallLLM_ExponentialBackoff tests that backoff delays are exponential
func TestCallLLM_ExponentialBackoff(t *testing.T) {
	startTime := time.Now()

	client := &Client{
		kind:  "openai-chat",
		model: "gpt-4-turbo",
		client: &http.Client{
			Transport: &mockRoundTripper{
				responses: []http.Response{
					{
						StatusCode: 500,
						Body:       io.NopCloser(bytes.NewBufferString("Error 1")),
						Header:     make(http.Header),
					},
					{
						StatusCode: 500,
						Body:       io.NopCloser(bytes.NewBufferString("Error 2")),
						Header:     make(http.Header),
					},
					{
						StatusCode: 500,
						Body:       io.NopCloser(bytes.NewBufferString("Error 3")),
						Header:     make(http.Header),
					},
					{
						StatusCode: 500,
						Body:       io.NopCloser(bytes.NewBufferString("Error 4")),
						Header:     make(http.Header),
					},
				},
			},
		},
	}

	ctx := context.Background()
	_, _ = client.callLLM(ctx, "system", "user")

	elapsed := time.Since(startTime)

	// Should have delays: 1s (after attempt 0) + 2s (after attempt 1) + 4s (after attempt 2) = 7 seconds minimum
	expectedMinimum := 7 * time.Second
	if elapsed < expectedMinimum {
		t.Errorf("Expected minimum elapsed time of %v, but got %v", expectedMinimum, elapsed)
	}

	// Allow some overhead (max 9 seconds to account for execution time)
	expectedMaximum := 9 * time.Second
	if elapsed > expectedMaximum {
		t.Errorf("Elapsed time %v exceeds expected maximum of %v", elapsed, expectedMaximum)
	}
}

// TestClient_CompletionAssessment_MockSuccess tests CompletionAssessment with mock mode (all AC passed)
func TestClient_CompletionAssessment_MockSuccess(t *testing.T) {
	client := NewMockClient()

	summary := &TaskSummary{
		Title: "Test Task",
		State: "RUNNING",
		AcceptanceCriteria: []AcceptanceCriterion{
			{ID: "AC-1", Description: "Feature A implemented", Type: "e2e", Critical: true},
			{ID: "AC-2", Description: "Feature B tested", Type: "unit", Critical: false},
		},
		WorkerRunsCount: 1,
		WorkerRuns: []WorkerRunSummary{
			{ID: "run-1", ExitCode: 0, Summary: "Success"},
		},
	}

	result, err := client.CompletionAssessment(context.Background(), summary)

	if err != nil {
		t.Fatalf("CompletionAssessment failed: %v", err)
	}

	if !result.AllCriteriaSatisfied {
		t.Errorf("Expected AllCriteriaSatisfied=true, got false")
	}

	if len(result.ByCriterion) != 2 {
		t.Errorf("Expected 2 CriterionResults, got %d", len(result.ByCriterion))
	}

	for _, cr := range result.ByCriterion {
		if cr.Status != "passed" {
			t.Errorf("Expected status=passed for %s, got %s", cr.ID, cr.Status)
		}
	}

	if result.Summary != "Mock: All criteria satisfied" {
		t.Errorf("Expected mock summary, got %q", result.Summary)
	}
}

// TestClient_CompletionAssessment_SomeFailed tests CompletionAssessment with partial failure
func TestClient_CompletionAssessment_SomeFailed(t *testing.T) {
	responseYAML := `type: completion_assessment
version: 1
payload:
  all_criteria_satisfied: false
  summary: "AC-2 not satisfied"
  by_criterion:
    - id: "AC-1"
      status: "passed"
      comment: "Feature A implemented correctly"
    - id: "AC-2"
      status: "failed"
      comment: "Test coverage insufficient"`

	// Create JSON-safe response
	jsonResponse := fmt.Sprintf(`{"choices":[{"message":{"role":"assistant","content":%q}}]}`, responseYAML)

	client := &Client{
		kind:  "openai-chat",
		model: "gpt-4-turbo",
		client: &http.Client{
			Transport: mockRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(jsonResponse)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	summary := &TaskSummary{
		Title: "Test Task",
		State: "RUNNING",
		AcceptanceCriteria: []AcceptanceCriterion{
			{ID: "AC-1", Description: "Feature A", Type: "e2e", Critical: true},
			{ID: "AC-2", Description: "Feature B", Type: "unit", Critical: true},
		},
		WorkerRunsCount: 1,
		WorkerRuns: []WorkerRunSummary{
			{ID: "run-1", ExitCode: 0, Summary: "Partial success"},
		},
	}

	result, err := client.CompletionAssessment(context.Background(), summary)

	if err != nil {
		t.Fatalf("CompletionAssessment failed: %v", err)
	}

	if result.AllCriteriaSatisfied {
		t.Errorf("Expected AllCriteriaSatisfied=false, got true")
	}

	if len(result.ByCriterion) != 2 {
		t.Errorf("Expected 2 CriterionResults, got %d", len(result.ByCriterion))
	}

	// Verify AC-1 passed
	if result.ByCriterion[0].ID != "AC-1" || result.ByCriterion[0].Status != "passed" {
		t.Errorf("Expected AC-1 passed, got %s with status %s", result.ByCriterion[0].ID, result.ByCriterion[0].Status)
	}

	// Verify AC-2 failed
	if result.ByCriterion[1].ID != "AC-2" || result.ByCriterion[1].Status != "failed" {
		t.Errorf("Expected AC-2 failed, got %s with status %s", result.ByCriterion[1].ID, result.ByCriterion[1].Status)
	}

	if result.Summary != "AC-2 not satisfied" {
		t.Errorf("Expected summary 'AC-2 not satisfied', got %q", result.Summary)
	}
}

// TestClient_CompletionAssessment_MarkdownCodeBlock tests YAML extraction from markdown code blocks
func TestClient_CompletionAssessment_MarkdownCodeBlock(t *testing.T) {
	responseMarkdown := "```yaml\ntype: completion_assessment\nversion: 1\npayload:\n  all_criteria_satisfied: true\n  summary: \"All criteria met\"\n  by_criterion:\n    - id: \"AC-1\"\n      status: \"passed\"\n      comment: \"Success\"\n```"

	// Create JSON-safe response
	jsonResponse := fmt.Sprintf(`{"choices":[{"message":{"role":"assistant","content":%q}}]}`, responseMarkdown)

	client := &Client{
		kind:  "openai-chat",
		model: "gpt-4-turbo",
		client: &http.Client{
			Transport: mockRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(jsonResponse)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	summary := &TaskSummary{
		Title: "Test Task",
		State: "RUNNING",
		AcceptanceCriteria: []AcceptanceCriterion{
			{ID: "AC-1", Description: "Feature A", Type: "e2e", Critical: true},
		},
		WorkerRunsCount: 1,
		WorkerRuns: []WorkerRunSummary{
			{ID: "run-1", ExitCode: 0, Summary: "Success"},
		},
	}

	result, err := client.CompletionAssessment(context.Background(), summary)

	if err != nil {
		t.Fatalf("CompletionAssessment failed with markdown code block: %v", err)
	}

	if !result.AllCriteriaSatisfied {
		t.Errorf("Expected AllCriteriaSatisfied=true, got false")
	}

	if len(result.ByCriterion) != 1 {
		t.Errorf("Expected 1 CriterionResult, got %d", len(result.ByCriterion))
	}

	if result.ByCriterion[0].Status != "passed" {
		t.Errorf("Expected status=passed, got %s", result.ByCriterion[0].Status)
	}
}

// TestClient_CompletionAssessment_RetryOn5xx tests retry behavior on 5xx errors
func TestClient_CompletionAssessment_RetryOn5xx(t *testing.T) {
	responseYAML := `type: completion_assessment
version: 1
payload:
  all_criteria_satisfied: true
  summary: "Recovered after retry"
  by_criterion:
    - id: "AC-1"
      status: "passed"
      comment: "Success after retry"`

	// Create JSON-safe response
	jsonResponse := fmt.Sprintf(`{"choices":[{"message":{"role":"assistant","content":%q}}]}`, responseYAML)

	client := &Client{
		kind:  "openai-chat",
		model: "gpt-4-turbo",
		client: &http.Client{
			Transport: &mockRoundTripper{
				responses: []http.Response{
					{
						StatusCode: 503,
						Body:       io.NopCloser(bytes.NewBufferString("Service Unavailable")),
						Header:     make(http.Header),
					},
					{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewBufferString(jsonResponse)),
						Header:     make(http.Header),
					},
				},
			},
		},
	}

	summary := &TaskSummary{
		Title: "Test Task",
		State: "RUNNING",
		AcceptanceCriteria: []AcceptanceCriterion{
			{ID: "AC-1", Description: "Feature A", Type: "e2e", Critical: true},
		},
		WorkerRunsCount: 1,
		WorkerRuns: []WorkerRunSummary{
			{ID: "run-1", ExitCode: 0, Summary: "Success"},
		},
	}

	result, err := client.CompletionAssessment(context.Background(), summary)

	if err != nil {
		t.Fatalf("CompletionAssessment failed after retry: %v", err)
	}

	if !result.AllCriteriaSatisfied {
		t.Errorf("Expected AllCriteriaSatisfied=true after retry, got false")
	}

	if result.Summary != "Recovered after retry" {
		t.Errorf("Expected summary 'Recovered after retry', got %q", result.Summary)
	}
}

// TestClient_CompletionAssessment_MultipleACs tests assessment with multiple acceptance criteria
func TestClient_CompletionAssessment_MultipleACs(t *testing.T) {
	responseYAML := `type: completion_assessment
version: 1
payload:
  all_criteria_satisfied: true
  summary: "All 5 acceptance criteria met"
  by_criterion:
    - id: "AC-1"
      status: "passed"
      comment: "Authentication implemented"
    - id: "AC-2"
      status: "passed"
      comment: "Authorization working"
    - id: "AC-3"
      status: "passed"
      comment: "Data validation complete"
    - id: "AC-4"
      status: "passed"
      comment: "Error handling verified"
    - id: "AC-5"
      status: "passed"
      comment: "Tests passing"`

	// Create JSON-safe response
	jsonResponse := fmt.Sprintf(`{"choices":[{"message":{"role":"assistant","content":%q}}]}`, responseYAML)

	client := &Client{
		kind:  "openai-chat",
		model: "gpt-4-turbo",
		client: &http.Client{
			Transport: mockRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(jsonResponse)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	summary := &TaskSummary{
		Title: "Complex Task",
		State: "RUNNING",
		AcceptanceCriteria: []AcceptanceCriterion{
			{ID: "AC-1", Description: "Authentication", Type: "e2e", Critical: true},
			{ID: "AC-2", Description: "Authorization", Type: "e2e", Critical: true},
			{ID: "AC-3", Description: "Data validation", Type: "unit", Critical: true},
			{ID: "AC-4", Description: "Error handling", Type: "unit", Critical: false},
			{ID: "AC-5", Description: "Tests", Type: "integration", Critical: true},
		},
		WorkerRunsCount: 2,
		WorkerRuns: []WorkerRunSummary{
			{ID: "run-1", ExitCode: 0, Summary: "Implementation complete"},
			{ID: "run-2", ExitCode: 0, Summary: "Tests added"},
		},
	}

	result, err := client.CompletionAssessment(context.Background(), summary)

	if err != nil {
		t.Fatalf("CompletionAssessment failed: %v", err)
	}

	if !result.AllCriteriaSatisfied {
		t.Errorf("Expected AllCriteriaSatisfied=true, got false")
	}

	if len(result.ByCriterion) != 5 {
		t.Errorf("Expected 5 CriterionResults, got %d", len(result.ByCriterion))
	}

	// Verify all criteria passed
	for i, cr := range result.ByCriterion {
		expectedID := fmt.Sprintf("AC-%d", i+1)
		if cr.ID != expectedID {
			t.Errorf("Expected ID=%s at index %d, got %s", expectedID, i, cr.ID)
		}
		if cr.Status != "passed" {
			t.Errorf("Expected status=passed for %s, got %s", cr.ID, cr.Status)
		}
		if cr.Comment == "" {
			t.Errorf("Expected non-empty comment for %s", cr.ID)
		}
	}

	if !strings.Contains(result.Summary, "All 5 acceptance criteria") {
		t.Errorf("Expected summary mentioning 5 criteria, got %q", result.Summary)
	}
}
