package mock

import (
	"github.com/biwakonbu/agent-runner/internal/core"
)

type NoteWriter struct {
	WriteFunc func(taskCtx *core.TaskContext) error
}

func (n *NoteWriter) Write(taskCtx *core.TaskContext) error {
	if n.WriteFunc != nil {
		return n.WriteFunc(taskCtx)
	}
	return nil
}

// NewMockNoteWriter creates a mock NoteWriter with default behavior
func NewMockNoteWriter() *NoteWriter {
	return &NoteWriter{}
}
