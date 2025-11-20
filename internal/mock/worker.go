package mock

import (
	"context"

	"github.com/biwakonbu/agent-runner/internal/core"
)

type WorkerExecutor struct {
	RunWorkerFunc func(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error)
}

func (w *WorkerExecutor) RunWorker(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error) {
	if w.RunWorkerFunc != nil {
		return w.RunWorkerFunc(ctx, prompt, env)
	}
	return nil, nil
}
