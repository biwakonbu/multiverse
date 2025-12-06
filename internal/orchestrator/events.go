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
	EventTaskStateChange      = "task:stateChange"
	EventExecutionStateChange = "execution:stateChange"
)

// TaskStateChangeEvent represents a task state change event
type TaskStateChangeEvent struct {
	TaskID    string     `json:"taskId"`
	OldStatus TaskStatus `json:"oldStatus"`
	NewStatus TaskStatus `json:"newStatus"`
	Timestamp time.Time  `json:"timestamp"`
}

// ExecutionStateChangeEvent represents an execution state change event
type ExecutionStateChangeEvent struct {
	OldState  ExecutionState `json:"oldState"`
	NewState  ExecutionState `json:"newState"`
	Timestamp time.Time      `json:"timestamp"`
}
