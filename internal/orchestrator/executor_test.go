package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestExecutor_ExecuteTask_Cancellation verifies that canceling the context kills the process.
func TestExecutor_ExecuteTask_Cancellation(t *testing.T) {
	// 1. Create a dummy agent-runner script
	tmpDir := t.TempDir()
	mockRunnerPath := filepath.Join(tmpDir, "mock_runner.sh")

	// Script using exec to replace shell process so kill works on it directly
	scriptContent := `#!/bin/sh
exec sleep 10
`
	err := os.WriteFile(mockRunnerPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("failed to write mock runner: %v", err)
	}

	// 2. Setup Executor (Stateless now, no TaskStore)
	executor := NewExecutor(mockRunnerPath, tmpDir)

	// 3. Create a Task (DTO)
	task := &Task{
		ID:     "task-cancel-test",
		Title:  "Sleep Task",
		Status: TaskStatusPending,
		PoolID: "default",
	}

	// 4. Execute with cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cleanup

	started := make(chan struct{})
	done := make(chan error)

	go func() {
		close(started)
		_, err := executor.ExecuteTask(ctx, task)
		done <- err
	}()

	<-started
	// Wait a bit to let the process start
	time.Sleep(500 * time.Millisecond)

	// 5. Cancel context
	cancel()

	// 6. Wait for result
	select {
	case err := <-done:
		// ExecuteTask returns error on cancellation
		assert.Error(t, err)
		t.Logf("ExecuteTask returned: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("ExecuteTask did not return after cancellation")
	}

	// 7. Verification:
	// Since Executor is stateless, we don't check TaskStore.
	// We check that it returned an error (asserted above).
	// In real usage, the Orchestrator calling this would handle saving Failed status.
}

// TestGenerateTaskYAML verifies that V2 fields are correctly correctly populated in the YAML
func TestGenerateTaskYAML(t *testing.T) {
	// 1. Setup Executor (mocking dependencies not needed for this method)
	executor := &Executor{}

	// 2. Create Task with V2 fields
	task := &Task{
		ID:                 "task-v2-test",
		Title:              "V2 Feature",
		Description:        "Implement V2 feature with AI",
		WBSLevel:           2,
		PhaseName:          "Implementation",
		Dependencies:       []string{"task-dep-1", "task-dep-2"},
		AcceptanceCriteria: []string{"AC1: works", "AC2: fast"},
		SuggestedImpl: &SuggestedImpl{
			Language:    "go",
			FilePaths:   []string{"main.go", "utils.go"},
			Constraints: []string{"No external libs", "Use stdlib"},
		},
	}

	// 3. Generate YAML
	yamlStr := executor.generateTaskYAML(task)

	// 4. Verify Content
	assert.Contains(t, yamlStr, `id: task-v2-test`)
	assert.Contains(t, yamlStr, `title: "V2 Feature"`)
	assert.Contains(t, yamlStr, `description: "Implement V2 feature with AI"`)
	assert.Contains(t, yamlStr, `wbs_level: 2`)
	assert.Contains(t, yamlStr, `phase_name: "Implementation"`)
	assert.Contains(t, yamlStr, `dependencies: ["task-dep-1", "task-dep-2"]`)

	// Check SuggestedImpl structured field
	assert.Contains(t, yamlStr, `suggested_impl:`)
	assert.Contains(t, yamlStr, `language: "go"`)
	assert.Contains(t, yamlStr, `file_paths: ["main.go", "utils.go"]`)
	assert.Contains(t, yamlStr, `constraints: ["No external libs", "Use stdlib"]`)

	// Check PRD Text includes legacy-compatible descriptions
	assert.Contains(t, yamlStr, "      Execute task: V2 Feature")
	assert.Contains(t, yamlStr, "      Description:")
	assert.Contains(t, yamlStr, "      Implement V2 feature with AI")
	assert.Contains(t, yamlStr, "      Acceptance Criteria:")
	assert.Contains(t, yamlStr, "      - AC1: works")
	assert.Contains(t, yamlStr, "      Suggested Implementation:")
	assert.Contains(t, yamlStr, "      Language: go")
}
