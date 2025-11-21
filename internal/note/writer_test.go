package note

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/biwakonbu/agent-runner/internal/core"
)

func TestWriter_Write_CreatesDotAgentRunnerDir(t *testing.T) {
	tmpDir := t.TempDir()

	ctx := &core.TaskContext{
		ID:         "TASK-001",
		Title:      "Test Task",
		RepoPath:   tmpDir,
		State:      core.StateComplete,
		PRDText:    "Sample PRD",
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
	}

	writer := NewWriter()
	err := writer.Write(ctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	agentRunnerDir := filepath.Join(tmpDir, ".agent-runner")
	if _, err := os.Stat(agentRunnerDir); os.IsNotExist(err) {
		t.Errorf(".agent-runner directory was not created")
	}
}

func TestWriter_Write_CreatesTaskNoteFile(t *testing.T) {
	tmpDir := t.TempDir()

	ctx := &core.TaskContext{
		ID:         "TASK-001",
		Title:      "Test Task",
		RepoPath:   tmpDir,
		State:      core.StateComplete,
		PRDText:    "Sample PRD",
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
	}

	writer := NewWriter()
	err := writer.Write(ctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	expectedFile := filepath.Join(tmpDir, ".agent-runner", "task-TASK-001.md")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("task-TASK-001.md was not created at %s", expectedFile)
	}
}

func TestWriter_Write_FileContainsTaskID(t *testing.T) {
	tmpDir := t.TempDir()

	ctx := &core.TaskContext{
		ID:         "TASK-002",
		Title:      "Test Task",
		RepoPath:   tmpDir,
		State:      core.StateComplete,
		PRDText:    "Sample PRD",
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
	}

	writer := NewWriter()
	err := writer.Write(ctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, ".agent-runner", "task-TASK-002.md"))
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	if !strings.Contains(string(content), "TASK-002") {
		t.Errorf("File does not contain task ID 'TASK-002'")
	}
}

func TestWriter_Write_FileContainsTitle(t *testing.T) {
	tmpDir := t.TempDir()

	taskTitle := "My Test Task Title"
	ctx := &core.TaskContext{
		ID:         "TASK-003",
		Title:      taskTitle,
		RepoPath:   tmpDir,
		State:      core.StateComplete,
		PRDText:    "Sample PRD",
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
	}

	writer := NewWriter()
	err := writer.Write(ctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, ".agent-runner", "task-TASK-003.md"))
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	if !strings.Contains(string(content), taskTitle) {
		t.Errorf("File does not contain title '%s'", taskTitle)
	}
}

func TestWriter_Write_FileContainsPRD(t *testing.T) {
	tmpDir := t.TempDir()

	prdText := "This is the PRD content that should appear in the file"
	ctx := &core.TaskContext{
		ID:         "TASK-004",
		Title:      "Test Task",
		RepoPath:   tmpDir,
		State:      core.StateComplete,
		PRDText:    prdText,
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
	}

	writer := NewWriter()
	err := writer.Write(ctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, ".agent-runner", "task-TASK-004.md"))
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	if !strings.Contains(string(content), prdText) {
		t.Errorf("File does not contain PRD text '%s'", prdText)
	}
}

func TestWriter_Write_FileContainsState(t *testing.T) {
	tmpDir := t.TempDir()

	ctx := &core.TaskContext{
		ID:         "TASK-005",
		Title:      "Test Task",
		RepoPath:   tmpDir,
		State:      core.StateFailed,
		PRDText:    "Sample PRD",
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
	}

	writer := NewWriter()
	err := writer.Write(ctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, ".agent-runner", "task-TASK-005.md"))
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	if !strings.Contains(string(content), "FAILED") {
		t.Errorf("File does not contain state 'FAILED'")
	}
}

func TestWriter_Write_WithAcceptanceCriteria(t *testing.T) {
	tmpDir := t.TempDir()

	ctx := &core.TaskContext{
		ID:       "TASK-006",
		Title:    "Test Task",
		RepoPath: tmpDir,
		State:    core.StateComplete,
		PRDText:  "Sample PRD",
		AcceptanceCriteria: []core.AcceptanceCriterion{
			{
				ID:          "AC-1",
				Description: "First criterion",
				Passed:      true,
			},
			{
				ID:          "AC-2",
				Description: "Second criterion",
				Passed:      false,
			},
		},
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
	}

	writer := NewWriter()
	err := writer.Write(ctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, ".agent-runner", "task-TASK-006.md"))
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "AC-1") {
		t.Errorf("File does not contain criterion ID 'AC-1'")
	}
	if !strings.Contains(contentStr, "First criterion") {
		t.Errorf("File does not contain criterion description 'First criterion'")
	}
	if !strings.Contains(contentStr, "AC-2") {
		t.Errorf("File does not contain criterion ID 'AC-2'")
	}
}

func TestWriter_Write_WithMetaCalls(t *testing.T) {
	tmpDir := t.TempDir()

	ctx := &core.TaskContext{
		ID:       "TASK-007",
		Title:    "Test Task",
		RepoPath: tmpDir,
		State:    core.StateComplete,
		PRDText:  "Sample PRD",
		MetaCalls: []core.MetaCallLog{
			{
				Type:         "plan_task",
				Timestamp:    time.Now(),
				RequestYAML:  "request: plan_task",
				ResponseYAML: "response: acceptance_criteria",
			},
		},
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
	}

	writer := NewWriter()
	err := writer.Write(ctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, ".agent-runner", "task-TASK-007.md"))
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "plan_task") {
		t.Errorf("File does not contain meta call type 'plan_task'")
	}
	if !strings.Contains(contentStr, "3.1 Meta Calls") {
		t.Errorf("File does not have Meta Calls section")
	}
}

func TestWriter_Write_WithWorkerRuns(t *testing.T) {
	tmpDir := t.TempDir()

	ctx := &core.TaskContext{
		ID:       "TASK-008",
		Title:    "Test Task",
		RepoPath: tmpDir,
		State:    core.StateComplete,
		PRDText:  "Sample PRD",
		WorkerRuns: []core.WorkerRunResult{
			{
				ID:         "run-1",
				StartedAt:  time.Now(),
				FinishedAt: time.Now(),
				ExitCode:   0,
				RawOutput:  "Worker output here",
				Summary:    "Worker executed successfully",
				Error:      nil,
			},
		},
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
	}

	writer := NewWriter()
	err := writer.Write(ctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, ".agent-runner", "task-TASK-008.md"))
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "run-1") {
		t.Errorf("File does not contain worker run ID 'run-1'")
	}
	if !strings.Contains(contentStr, "3.2 Worker Runs") {
		t.Errorf("File does not have Worker Runs section")
	}
	if !strings.Contains(contentStr, "Worker output here") {
		t.Errorf("File does not contain worker output")
	}
}

func TestWriter_Write_EmptyTaskContext(t *testing.T) {
	tmpDir := t.TempDir()

	ctx := &core.TaskContext{
		ID:       "EMPTY-001",
		RepoPath: tmpDir,
	}

	writer := NewWriter()
	err := writer.Write(ctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	expectedFile := filepath.Join(tmpDir, ".agent-runner", "task-EMPTY-001.md")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("task-EMPTY-001.md was not created")
	}
}

func TestWriter_Write_WithNilFields(t *testing.T) {
	tmpDir := t.TempDir()

	ctx := &core.TaskContext{
		ID:       "TASK-009",
		Title:    "Test",
		RepoPath: tmpDir,
		State:    core.StateComplete,
		// All other fields are nil/empty
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
	}

	writer := NewWriter()
	err := writer.Write(ctx)
	if err != nil {
		t.Fatalf("Write() with nil fields error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, ".agent-runner", "task-TASK-009.md"))
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	if len(content) == 0 {
		t.Errorf("Generated file is empty")
	}
}

func TestWriter_Write_TemplateCorrectness(t *testing.T) {
	tmpDir := t.TempDir()

	ctx := &core.TaskContext{
		ID:       "TASK-010",
		Title:    "Template Test",
		RepoPath: tmpDir,
		State:    core.StateComplete,
		PRDText:  "PRD Content",
		AcceptanceCriteria: []core.AcceptanceCriterion{
			{
				ID:          "AC-1",
				Description: "Test AC",
				Passed:      true,
			},
		},
		StartedAt:  time.Now(),
		FinishedAt: time.Now(),
	}

	writer := NewWriter()
	err := writer.Write(ctx)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, ".agent-runner", "task-TASK-010.md"))
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Check markdown structure
	sections := []string{
		"# Task Note",
		"## 1. PRD Summary",
		"## 2. Acceptance Criteria",
		"## 3. Execution Log",
	}

	for _, section := range sections {
		if !strings.Contains(contentStr, section) {
			t.Errorf("File does not contain section '%s'", section)
		}
	}
}
