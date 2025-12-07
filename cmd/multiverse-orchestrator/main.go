package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/biwakonbu/agent-runner/internal/orchestrator"
	"github.com/biwakonbu/agent-runner/internal/orchestrator/ipc"
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
	taskStore := orchestrator.NewTaskStore(*workspaceDir)
	queue := ipc.NewFilesystemQueue(*workspaceDir)
	executor := orchestrator.NewExecutor(*agentRunnerPath, *workspaceDir, taskStore)

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	log.Printf("Orchestrator started. Workspace: %s, Pool: %s", *workspaceDir, *poolID)

	// Polling loop
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Attempt to dequeue a job
			job, err := queue.Dequeue(*poolID)
			if err != nil {
				log.Printf("Error checking queue: %v", err)
				continue
			}

			if job == nil {
				// No jobs
				continue
			}

			// Process job
			if err := processJob(ctx, executor, queue, job); err != nil {
				log.Printf("Failed to process job %s: %v", job.ID, err)
			}
		}
	}
}

func processJob(ctx context.Context, executor *orchestrator.Executor, queue *ipc.FilesystemQueue, job *ipc.Job) error {
	log.Printf("Processing job: %s (Task: %s)", job.ID, job.TaskID)

	// Load Task
	task, err := executor.TaskStore.LoadTask(job.TaskID)
	if err != nil {
		return fmt.Errorf("failed to load task %s: %w", job.TaskID, err)
	}

	// Execute
	attempt, err := executor.ExecuteTask(ctx, task)
	if err != nil {
		log.Printf("Task execution failed: %v", err)
	}

	if attempt != nil {
		log.Printf("Task finished. Status: %s, ID: %s", attempt.Status, attempt.ID)
	}

	// Complete job (remove from processing)
	if err := queue.Complete(job.ID, job.PoolID); err != nil {
		return fmt.Errorf("failed to complete job: %w", err)
	}

	return nil
}
