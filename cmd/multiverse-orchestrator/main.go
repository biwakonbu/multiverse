package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/biwakonbu/agent-runner/internal/orchestrator"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/ipc"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/persistence"
)

func main() {
	// Parse flags
	workspaceDir := flag.String("workspace", filepath.Join(os.Getenv("HOME"), ".multiverse"), "Path to multiverse workspace directory")
	agentRunnerPath := flag.String("agent-runner", "agent-runner", "Path to agent-runner binary")
	poolID := flag.String("pool", "default", "Queue Pool ID to consume from")
	flag.Parse()

	// Validate workspace
	if _, err := os.Stat(*workspaceDir); os.IsNotExist(err) {
		log.Fatalf("Workspace directory does not exist: %s", *workspaceDir)
	}

	// Initialize components
	repo := persistence.NewWorkspaceRepository(*workspaceDir)
	if err := repo.Init(); err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}

	queue := ipc.NewFilesystemQueue(*workspaceDir)

	// Scheduler (Optional for pure worker, but Orchestrator usually bundles both roles in this binary?)
	// If this binary acts as the Orchestrator Daemon, it should process schedule + execution.
	scheduler := orchestrator.NewScheduler(repo, queue, nil)

	// Executor (Stateless)
	executor := orchestrator.NewExecutor(*agentRunnerPath, *workspaceDir)

	// RetryPolicy and Backlog configurable? Using defaults for now.
	backlogStore := orchestrator.NewBacklogStore(*workspaceDir)

	// Create ExecutionOrchestrator
	orch := orchestrator.NewExecutionOrchestrator(
		scheduler,
		executor,
		repo,
		queue,
		nil, // EventEmitter (can be added if needed for logging/UI)
		backlogStore,
		[]string{*poolID},
	)

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start Orchestrator
	log.Printf("Orchestrator started. Workspace: %s, Pool: %s", *workspaceDir, *poolID)
	if err := orch.Start(ctx); err != nil {
		log.Fatalf("Failed to start orchestrator: %v", err)
	}

	// Wait for signal
	<-sigChan
	log.Println("Shutting down...")

	// Stop gracefully
	if err := orch.Stop(); err != nil {
		log.Printf("Error stopping orchestrator: %v", err)
	}
	orch.Wait()
	log.Println("Orchestrator stopped.")
}
