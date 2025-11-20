package core

import (
	"time"

	"github.com/biwakonbu/agent-runner/pkg/config"
)

// TaskState represents the current state of the task FSM
type TaskState string

const (
	StatePending    TaskState = "PENDING"
	StatePlanning   TaskState = "PLANNING"
	StateRunning    TaskState = "RUNNING"
	StateValidating TaskState = "VALIDATING"
	StateComplete   TaskState = "COMPLETE"
	StateFailed     TaskState = "FAILED"
)

// TaskContext holds the state of the current task
type TaskContext struct {
	ID       string
	Title    string
	RepoPath string
	State    TaskState

	PRDText string

	AcceptanceCriteria []AcceptanceCriterion
	MetaCalls          []MetaCallLog
	WorkerRuns         []WorkerRunResult

	TestConfig *config.TestDetails
	TestResult *TestResult

	StartedAt  time.Time
	FinishedAt time.Time
}

// AcceptanceCriterion represents a single acceptance criterion from Meta
type AcceptanceCriterion struct {
	ID          string `yaml:"id"`
	Description string `yaml:"description"`
	Passed      bool
}

// MetaCallLog records a request/response pair with Meta
type MetaCallLog struct {
	Type         string
	Timestamp    time.Time
	RequestYAML  string
	ResponseYAML string
}

// WorkerRunResult records a single execution of the worker
type WorkerRunResult struct {
	ID         string
	StartedAt  time.Time
	FinishedAt time.Time
	ExitCode   int
	RawOutput  string
	Summary    string
	Error      error
}

// TestResult records the result of the test command
type TestResult struct {
	Command   string
	ExitCode  int
	Summary   string
	RawOutput string
}
