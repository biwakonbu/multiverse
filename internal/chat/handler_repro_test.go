package chat_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/biwakonbu/agent-runner/internal/chat"
	"github.com/biwakonbu/agent-runner/internal/meta"
	"github.com/biwakonbu/agent-runner/internal/orchestrator"
)

func TestHandleMessage_Mock(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	taskStore := orchestrator.NewTaskStore(tmpDir)
	sessionStore := chat.NewChatSessionStore(tmpDir)
	mockMeta := meta.NewMockClient()

	// Create Handler
	handler := chat.NewHandler(mockMeta, taskStore, sessionStore, "test-ws", tmpDir, nil)

	// Create Session
	ctx := context.Background()
	session, err := handler.CreateSession(ctx)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Send Message
	resp, err := handler.HandleMessage(ctx, session.ID, "Make a todo app")
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	// Verify Response
	if resp.Message.Content == "" {
		t.Error("Expected non-empty response content")
	}
	if len(resp.GeneratedTasks) == 0 {
		t.Error("Expected generated tasks from mock")
	}

	// Verify History (User + Assistant)
	history, err := handler.GetHistory(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	// 1 system + 1 user + 1 assistant = 3 messages
	if len(history) != 3 {
		t.Errorf("Expected 3 messages in history, got %d", len(history))
		for i, msg := range history {
			t.Logf("[%d] %s: %s", i, msg.Role, msg.Content)
		}
	}

	if history[1].Role != "user" {
		t.Errorf("Expected second message to be user, got %s", history[1].Role)
	}
	if history[2].Role != "assistant" {
		t.Errorf("Expected third message to be assistant, got %s", history[2].Role)
	}
}

func TestHandleMessage_EmitsFailedEventOnMetaError(t *testing.T) {
	tmpDir := t.TempDir()
	taskStore := orchestrator.NewTaskStore(tmpDir)
	sessionStore := chat.NewChatSessionStore(tmpDir)
	recorder := &recordingEmitter{}
	handler := chat.NewHandler(failingMetaClient{}, taskStore, sessionStore, "ws", tmpDir, recorder)

	ctx := context.Background()
	session, err := handler.CreateSession(ctx)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if _, err := handler.HandleMessage(ctx, session.ID, "hi"); err == nil {
		t.Fatalf("expected error but got nil")
	}

	if len(recorder.events) == 0 {
		t.Fatalf("expected progress events to be emitted")
	}
	last := recorder.events[len(recorder.events)-1]
	if last.Step != "Failed" {
		t.Fatalf("expected last event to be Failed, got %s", last.Step)
	}
}

func TestHandleMessage_FailsOnUnknownDependency(t *testing.T) {
	tmpDir := t.TempDir()
	taskStore := orchestrator.NewTaskStore(tmpDir)
	sessionStore := chat.NewChatSessionStore(tmpDir)
	recorder := &recordingEmitter{}

	resp := &meta.DecomposeResponse{
		Understanding: "Test",
		Phases: []meta.DecomposedPhase{
			{
				Name: "Impl",
				Tasks: []meta.DecomposedTask{
					{
						ID:           "t1",
						Title:        "Task1",
						Description:  "desc",
						Dependencies: []string{"ghost-task"},
						WBSLevel:     3,
					},
				},
			},
		},
	}

	handler := chat.NewHandler(staticMetaClient{resp: resp}, taskStore, sessionStore, "ws", tmpDir, recorder)

	ctx := context.Background()
	session, err := handler.CreateSession(ctx)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if _, err := handler.HandleMessage(ctx, session.ID, "hi"); err == nil {
		t.Fatalf("expected error but got nil")
	}

	tasks, err := taskStore.ListAllTasks()
	if err != nil {
		t.Fatalf("ListAllTasks failed: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected no tasks to be persisted on failure, got %d", len(tasks))
	}

	if len(recorder.events) == 0 {
		t.Fatalf("expected progress events to be emitted")
	}
	if recorder.events[len(recorder.events)-1].Step != "Failed" {
		t.Fatalf("expected last event Failed, got %s", recorder.events[len(recorder.events)-1].Step)
	}
}

type failingMetaClient struct{}

func (failingMetaClient) Decompose(context.Context, *meta.DecomposeRequest) (*meta.DecomposeResponse, error) {
	return nil, fmt.Errorf("meta failure")
}

type staticMetaClient struct {
	resp *meta.DecomposeResponse
}

func (s staticMetaClient) Decompose(context.Context, *meta.DecomposeRequest) (*meta.DecomposeResponse, error) {
	return s.resp, nil
}

type recordingEmitter struct {
	events []orchestrator.ChatProgressEvent
}

func (r *recordingEmitter) Emit(eventName string, data any) {
	if eventName != orchestrator.EventChatProgress {
		return
	}
	if ev, ok := data.(orchestrator.ChatProgressEvent); ok {
		r.events = append(r.events, ev)
	}
}
