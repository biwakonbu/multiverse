package orchestrator

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TaskStatus represents the status of a task.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "PENDING"
	TaskStatusReady     TaskStatus = "READY"
	TaskStatusRunning   TaskStatus = "RUNNING"
	TaskStatusSucceeded TaskStatus = "SUCCEEDED"
	TaskStatusCompleted TaskStatus = "COMPLETED"
	TaskStatusFailed    TaskStatus = "FAILED"
	TaskStatusCanceled  TaskStatus = "CANCELED"
	TaskStatusBlocked   TaskStatus = "BLOCKED"
	TaskStatusRetryWait TaskStatus = "RETRY_WAIT"
)

// Task represents a unit of work.
type Task struct {
	// 基本フィールド
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Status    TaskStatus `json:"status"`
	PoolID    string     `json:"poolId"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	StartedAt *time.Time `json:"startedAt,omitempty"`
	DoneAt    *time.Time `json:"doneAt,omitempty"`

	// v2.0 拡張フィールド: タスク分解・依存関係
	Description        string   `json:"description,omitempty"`        // タスクの詳細説明
	Dependencies       []string `json:"dependencies,omitempty"`       // 依存タスクIDリスト
	ParentID           *string  `json:"parentId,omitempty"`           // 親タスクID（WBS階層用）
	WBSLevel           int      `json:"wbsLevel,omitempty"`           // WBS階層レベル（1=概念設計, 2=実装設計, 3=実装）
	PhaseName          string   `json:"phaseName,omitempty"`          // フェーズ名
	Milestone          string   `json:"milestone,omitempty"`          // マイルストーン名（Phase単位のまとまり）
	SourceChatID       *string  `json:"sourceChatId,omitempty"`       // 生成元チャットセッションID
	AcceptanceCriteria []string `json:"acceptanceCriteria,omitempty"` // 達成条件リスト

	// リトライ管理用 (v2.0 Extension)
	AttemptCount int        `json:"attemptCount,omitempty"` // 試行回数
	NextRetryAt  *time.Time `json:"nextRetryAt,omitempty"`  // 次回リトライ予定時刻
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
	store := &TaskStore{WorkspaceDir: workspaceDir}
	// Ensure directories exist
	if err := os.MkdirAll(store.GetTaskDir(), 0755); err != nil {
		fmt.Printf("failed to create task dir: %v\n", err)
	}
	if err := os.MkdirAll(store.GetAttemptDir(), 0755); err != nil {
		fmt.Printf("failed to create attempt dir: %v\n", err)
	}
	return store
}

// GetTaskDir returns the directory for tasks.
func (s *TaskStore) GetTaskDir() string {
	return filepath.Join(s.WorkspaceDir, "tasks")
}

// GetAttemptDir returns the directory for attempts.
func (s *TaskStore) GetAttemptDir() string {
	return filepath.Join(s.WorkspaceDir, "attempts")
}

// ensureSafeID validates IDs to prevent path traversal or absolute paths.
func ensureSafeID(id string) error {
	if id == "" {
		return fmt.Errorf("id is empty")
	}
	if filepath.IsAbs(id) {
		return fmt.Errorf("absolute id is not allowed")
	}
	if strings.Contains(id, "..") || strings.ContainsAny(id, `/\ `) {
		return fmt.Errorf("id contains invalid path characters")
	}
	return nil
}

// LoadTask loads the latest state of a task by ID.
// It reads the last line of the JSONL file.
func (s *TaskStore) LoadTask(id string) (*Task, error) {
	if err := ensureSafeID(id); err != nil {
		return nil, err
	}

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
	if err := ensureSafeID(task.ID); err != nil {
		return err
	}

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
	if err := ensureSafeID(id); err != nil {
		return nil, err
	}

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
	if err := ensureSafeID(attempt.ID); err != nil {
		return err
	}
	if err := ensureSafeID(attempt.TaskID); err != nil {
		return fmt.Errorf("invalid task id: %w", err)
	}

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

// ListAttemptsByTaskID returns all attempts for a given task ID.
func (s *TaskStore) ListAttemptsByTaskID(taskID string) ([]Attempt, error) {
	dir := s.GetAttemptDir()
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return []Attempt{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read attempts directory: %w", err)
	}

	var attempts []Attempt
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		attempt, err := s.LoadAttempt(entry.Name()[:len(entry.Name())-5])
		if err != nil {
			continue
		}

		if attempt.TaskID == taskID {
			attempts = append(attempts, *attempt)
		}
	}

	return attempts, nil
}

// PoolSummary represents task counts by status for a pool.
type PoolSummary struct {
	PoolID  string         `json:"poolId"`
	Running int            `json:"running"`
	Queued  int            `json:"queued"`
	Failed  int            `json:"failed"`
	Total   int            `json:"total"`
	Counts  map[string]int `json:"counts"`
}

// GetPoolSummaries returns task count summaries by pool.
func (s *TaskStore) GetPoolSummaries() ([]PoolSummary, error) {
	dir := s.GetTaskDir()
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return []PoolSummary{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks directory: %w", err)
	}

	// poolID -> status -> count
	poolCounts := make(map[string]map[TaskStatus]int)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}

		id := entry.Name()[:len(entry.Name())-6]
		task, err := s.LoadTask(id)
		if err != nil {
			continue
		}

		if poolCounts[task.PoolID] == nil {
			poolCounts[task.PoolID] = make(map[TaskStatus]int)
		}
		poolCounts[task.PoolID][task.Status]++
	}

	// Convert to PoolSummary slice
	var summaries []PoolSummary
	for poolID, statusCounts := range poolCounts {
		summary := PoolSummary{
			PoolID:  poolID,
			Running: statusCounts[TaskStatusRunning],
			Queued:  statusCounts[TaskStatusPending] + statusCounts[TaskStatusReady],
			Failed:  statusCounts[TaskStatusFailed],
			Counts:  make(map[string]int),
		}
		for status, count := range statusCounts {
			summary.Counts[string(status)] = count
			summary.Total += count
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// Pool represents a worker pool configuration.
type Pool struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// DefaultPools はデフォルトの Pool 定義を返す
// 将来的には worker-pools.json から読み込む
var DefaultPools = []Pool{
	{ID: "default", Name: "Default", Description: "汎用タスク実行用"},
	{ID: "codegen", Name: "Codegen", Description: "コード生成タスク用"},
	{ID: "test", Name: "Test", Description: "テスト実行タスク用"},
}

// GetAvailablePools は利用可能な Pool 一覧を返す
func (s *TaskStore) GetAvailablePools() []Pool {
	path := filepath.Join(s.WorkspaceDir, "worker-pools.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultPools
	}

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("failed to read worker-pools.json: %v\n", err)
		return DefaultPools
	}

	var config struct {
		Pools []Pool `json:"pools"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("failed to parse worker-pools.json: %v\n", err)
		return DefaultPools
	}

	if len(config.Pools) == 0 {
		return DefaultPools
	}

	return config.Pools
}

// ListAllTasks は全タスクの最新状態を返す
func (s *TaskStore) ListAllTasks() ([]Task, error) {
	dir := s.GetTaskDir()
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return []Task{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks directory: %w", err)
	}

	var tasks []Task
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}

		id := entry.Name()[:len(entry.Name())-6]
		task, err := s.LoadTask(id)
		if err != nil {
			continue
		}
		tasks = append(tasks, *task)
	}

	return tasks, nil
}

// ListTasksByStatus は指定ステータスのタスク一覧を返す
func (s *TaskStore) ListTasksByStatus(status TaskStatus) ([]Task, error) {
	allTasks, err := s.ListAllTasks()
	if err != nil {
		return nil, err
	}

	var filtered []Task
	for _, task := range allTasks {
		if task.Status == status {
			filtered = append(filtered, task)
		}
	}
	return filtered, nil
}

// ListTasksBySourceChat は指定チャットセッションから生成されたタスク一覧を返す
func (s *TaskStore) ListTasksBySourceChat(chatID string) ([]Task, error) {
	allTasks, err := s.ListAllTasks()
	if err != nil {
		return nil, err
	}

	var filtered []Task
	for _, task := range allTasks {
		if task.SourceChatID != nil && *task.SourceChatID == chatID {
			filtered = append(filtered, task)
		}
	}
	return filtered, nil
}
