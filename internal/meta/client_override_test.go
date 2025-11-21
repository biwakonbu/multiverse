package meta

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// TestClient_SystemPromptOverride_UsedInRequest verifies that the override is actually sent to the LLM
func TestClient_SystemPromptOverride_UsedInRequest(t *testing.T) {
	override := "You are a custom agent."

	// Capture the request body
	var capturedRequest chatRequest

	mockTransport := mockRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		// Read body
		bodyBytes, _ := io.ReadAll(req.Body)
		req.Body.Close()                                    // Close original body
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore for any other readers (though not needed here)

		if err := json.Unmarshal(bodyBytes, &capturedRequest); err != nil {
			return nil, err
		}

		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(`{"choices":[{"message":{"role":"assistant","content":"type: plan_task\nversion: 1\npayload:\n  task_id: \"TASK-1\""}}]}`)),
			Header:     make(http.Header),
		}, nil
	})

	client := NewClient("openai-chat", "key", "gpt-4-turbo", override)
	client.client.Transport = mockTransport

	// Call PlanTask
	_, err := client.PlanTask(context.Background(), "test prd")
	if err != nil {
		t.Fatalf("PlanTask failed: %v", err)
	}

	// Verify system prompt in request
	foundSystem := false
	for _, msg := range capturedRequest.Messages {
		if msg.Role == "system" {
			foundSystem = true
			if msg.Content != override {
				t.Errorf("Expected system prompt %q, got %q", override, msg.Content)
			}
		}
	}

	if !foundSystem {
		t.Error("System prompt not found in request")
	}
}
