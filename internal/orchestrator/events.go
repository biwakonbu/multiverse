package orchestrator

import (
	"context"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// EventEmitter defines the interface for emitting events
type EventEmitter interface {
	Emit(eventName string, data any)
}

// WailsEventEmitter implements EventEmitter using Wails runtime
type WailsEventEmitter struct {
	ctx context.Context
}

// NewWailsEventEmitter creates a new WailsEventEmitter
func NewWailsEventEmitter(ctx context.Context) *WailsEventEmitter {
	return &WailsEventEmitter{ctx: ctx}
}

// Emit emits an event via Wails runtime
func (w *WailsEventEmitter) Emit(eventName string, data any) {
	if w.ctx != nil {
		runtime.EventsEmit(w.ctx, eventName, data)
	}
}

// Event names
const (
	EventTaskStateChange        = "task:stateChange"
	EventExecutionStateChange   = "execution:stateChange"
	EventTaskCreated            = "task:created"
	EventChatProgress           = "chat:progress"
	EventBacklogAdded           = "backlog:added"
	EventTaskLog                = "task:log"
	EventProcessMetaUpdate      = "process:metaUpdate"
	EventProcessWorkerUpdate    = "process:workerUpdate"
	EventProcessContainerUpdate = "process:containerUpdate"
)

// TaskStateChangeEvent represents a task state change event
type TaskStateChangeEvent struct {
	TaskID    string     `json:"taskId"`
	OldStatus TaskStatus `json:"oldStatus"`
	NewStatus TaskStatus `json:"newStatus"`
	Timestamp time.Time  `json:"timestamp"`
}

// TaskCreatedEvent represents a task creation event
type TaskCreatedEvent struct {
	Task Task `json:"task"`
}

// ExecutionStateChangeEvent represents an execution state change event
type ExecutionStateChangeEvent struct {
	OldState  ExecutionState `json:"oldState"`
	NewState  ExecutionState `json:"newState"`
	Timestamp time.Time      `json:"timestamp"`
}

// ChatProgressEvent represents a progress update during chat processing
type ChatProgressEvent struct {
	SessionID string    `json:"sessionId"`
	Step      string    `json:"step"`    // e.g. "Decomposing", "Persisting"
	Message   string    `json:"message"` // Human readable message
	Timestamp time.Time `json:"timestamp"`
}

// TaskLogEvent represents a log line from task execution
type TaskLogEvent struct {
	TaskID    string    `json:"taskId"`
	Stream    string    `json:"stream"` // "stdout" or "stderr"
	Line      string    `json:"line"`   // Log line content
	Timestamp time.Time `json:"timestamp"`
}

// ProcessMetaUpdateEvent represents a meta-agent state update
type ProcessMetaUpdateEvent struct {
	TaskID    string    `json:"taskId"`
	TaskTitle string    `json:"taskTitle,omitempty"`
	State     string    `json:"state"`  // e.g. "THINKING", "PLANNING", "ACTING", "OBSERVING"
	Detail    string    `json:"detail"` // e.g. "Analyzing dependencies..."
	Timestamp time.Time `json:"timestamp"`
}

// ProcessWorkerUpdateEvent represents a worker execution update
type ProcessWorkerUpdateEvent struct {
	TaskID    string    `json:"taskId"`
	WorkerID  string    `json:"workerId"`
	Status    string    `json:"status"` // "RUNNING", "IDLE", "ERROR"
	Command   string    `json:"command"`
	ExitCode  int       `json:"exitCode,omitempty"`
	Artifacts []string  `json:"artifacts,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ProcessContainerUpdateEvent represents a container lifecycle update
type ProcessContainerUpdateEvent struct {
	TaskID      string    `json:"taskId"`
	ContainerID string    `json:"containerId"`
	Status      string    `json:"status"` // "STARTING", "RUNNING", "STOPPED"
	Image       string    `json:"image"`
	Timestamp   time.Time `json:"timestamp"`
}
