package orchestrator

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TaskStatus represents the status of a task.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "PENDING"
	TaskStatusReady     TaskStatus = "READY"
	TaskStatusRunning   TaskStatus = "RUNNING"
	TaskStatusSucceeded TaskStatus = "SUCCEEDED"
	TaskStatusFailed    TaskStatus = "FAILED"
	TaskStatusCanceled  TaskStatus = "CANCELED"
	TaskStatusBlocked   TaskStatus = "BLOCKED"
)

// Task represents a unit of work.
type Task struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Status    TaskStatus `json:"status"`
	PoolID    string     `json:"poolId"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	StartedAt *time.Time `json:"startedAt,omitempty"`
	DoneAt    *time.Time `json:"doneAt,omitempty"`
}

// AttemptStatus represents the status of an attempt.
type AttemptStatus string

const (
	AttemptStatusStarting  AttemptStatus = "STARTING"
	AttemptStatusRunning   AttemptStatus = "RUNNING"
	AttemptStatusSucceeded AttemptStatus = "SUCCEEDED"
	AttemptStatusFailed    AttemptStatus = "FAILED"
	AttemptStatusTimeout   AttemptStatus = "TIMEOUT"
	AttemptStatusCanceled  AttemptStatus = "CANCELED"
)

// Attempt represents a single execution attempt of a task.
type Attempt struct {
	ID           string        `json:"id"`
	TaskID       string        `json:"taskId"`
	Status       AttemptStatus `json:"status"`
	StartedAt    time.Time     `json:"startedAt"`
	FinishedAt   *time.Time    `json:"finishedAt,omitempty"`
	ErrorSummary string        `json:"errorSummary,omitempty"`
}

// TaskStore handles task and attempt persistence.
type TaskStore struct {
	WorkspaceDir string
}

// NewTaskStore creates a new TaskStore for a specific workspace.
func NewTaskStore(workspaceDir string) *TaskStore {
	return &TaskStore{WorkspaceDir: workspaceDir}
}

// GetTaskDir returns the directory for tasks.
func (s *TaskStore) GetTaskDir() string {
	return filepath.Join(s.WorkspaceDir, "tasks")
}

// GetAttemptDir returns the directory for attempts.
func (s *TaskStore) GetAttemptDir() string {
	return filepath.Join(s.WorkspaceDir, "attempts")
}

// LoadTask loads the latest state of a task by ID.
// It reads the last line of the JSONL file.
func (s *TaskStore) LoadTask(id string) (*Task, error) {
	path := filepath.Join(s.GetTaskDir(), id+".jsonl")
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var lastLine string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			lastLine = line
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if lastLine == "" {
		return nil, fmt.Errorf("task file is empty: %s", id)
	}

	var task Task
	if err := json.Unmarshal([]byte(lastLine), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task line: %w", err)
	}

	return &task, nil
}

// SaveTask appends a new state of the task to the JSONL file.
func (s *TaskStore) SaveTask(task *Task) error {
	dir := s.GetTaskDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create tasks directory: %w", err)
	}

	path := filepath.Join(dir, task.ID+".jsonl")

	// Ensure UpdatedAt is set
	task.UpdatedAt = time.Now()

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open task file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write task line: %w", err)
	}

	return nil
}

// LoadAttempt loads an attempt by ID.
func (s *TaskStore) LoadAttempt(id string) (*Attempt, error) {
	path := filepath.Join(s.GetAttemptDir(), id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var attempt Attempt
	if err := json.Unmarshal(data, &attempt); err != nil {
		return nil, fmt.Errorf("failed to unmarshal attempt: %w", err)
	}

	return &attempt, nil
}

// SaveAttempt saves an attempt.
func (s *TaskStore) SaveAttempt(attempt *Attempt) error {
	dir := s.GetAttemptDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create attempts directory: %w", err)
	}

	path := filepath.Join(dir, attempt.ID+".json")
	data, err := json.MarshalIndent(attempt, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal attempt: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write attempt file: %w", err)
	}

	return nil
}
