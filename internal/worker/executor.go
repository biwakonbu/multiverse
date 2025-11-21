package worker

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/biwakonbu/agent-runner/internal/core"
	"github.com/biwakonbu/agent-runner/pkg/config"
)

type Executor struct {
	Config   config.WorkerConfig
	Sandbox  SandboxProvider
	RepoPath string
}

func NewExecutor(cfg config.WorkerConfig, repoPath string) (*Executor, error) {
	sb, err := NewSandboxManager()
	if err != nil {
		return nil, err
	}
	return &Executor{
		Config:   cfg,
		Sandbox:  sb,
		RepoPath: repoPath,
	}, nil
}

func (e *Executor) RunWorker(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error) {
	// 1. Start Sandbox (if not already running? For MVP, start new one per task or reuse?)
	// The Runner creates Executor once.
	// Ideally, Sandbox should be managed by Runner or Executor should manage lifecycle.
	// For MVP, let's start a container for the duration of RunWorker?
	// No, RunWorker is called multiple times in a loop. We want persistence.
	// But Executor interface is stateless `RunWorker`.

	// Let's assume we start a container for the whole task.
	// But `RunWorker` signature doesn't pass container ID.
	// We need to change the architecture slightly or manage state in Executor.

	// Let's make Executor stateful for the task.
	// Or, just start/stop for each run (slow, but simple for MVP).
	// The architecture diagram says "1 Task = 1 sandbox".
	// So the Sandbox should be started at the beginning of the task.

	// But `Runner` doesn't explicitly start sandbox.
	// Let's modify `Runner` or `Executor` to handle this.
	// `Executor` can hold the container ID.

	// NOTE: In a real app, we'd have `Start()` and `Stop()` methods on Executor.
	// For now, let's just start on first run and keep it?
	// Or start/stop every time (inefficient but safe).
	// Let's try start/stop every time for MVP simplicity,
	// BUT persistence of files is needed!
	// "Repo mount" handles persistence of files.
	// So start/stop is fine for file persistence, but we lose process state.
	// That's acceptable for "CLI" workers usually.

	// Let's go with Start/Stop per RunWorker for now to avoid complex lifecycle management in v1.

	image := e.Config.DockerImage
	if image == "" {
		image = "ghcr.io/biwakonbu/agent-runner-codex:latest" // Default
	}

	// Start Container
	// We need repoPath. I'll add it to Executor struct and set it during init.
	// But `NewExecutor` needs it.

	// For now, let's hardcode "." or use a field.
	repoPath := e.RepoPath
	if repoPath == "" {
		absRepo, err := filepath.Abs(".")
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path: %w", err)
		}
		repoPath = absRepo
	}

	containerID, err := e.Sandbox.StartContainer(ctx, image, repoPath, env)
	if err != nil {
		return nil, fmt.Errorf("failed to start sandbox: %w", err)
	}
	defer func() {
		_ = e.Sandbox.StopContainer(ctx, containerID)
	}()

	// Construct Command
	// "codex exec --sandbox workspace-write --json --cwd /workspace/project '<prompt>'"
	cmd := []string{
		"codex", "exec",
		"--sandbox", "workspace-write",
		"--json",
		"--cwd", "/workspace/project",
		prompt,
	}

	// Exec
	start := time.Now()
	exitCode, output, err := e.Sandbox.Exec(ctx, containerID, cmd)
	finish := time.Now()

	res := &core.WorkerRunResult{
		ID:         fmt.Sprintf("run-%d", start.Unix()),
		StartedAt:  start,
		FinishedAt: finish,
		ExitCode:   exitCode,
		RawOutput:  output,
		Summary:    "Worker executed", // Could parse JSON output if needed
		Error:      err,
	}

	return res, nil
}

// Add RepoPath to struct
// I need to update the struct definition above.
