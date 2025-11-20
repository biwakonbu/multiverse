package meta

// Protocol definitions for Meta-agent communication

// Common wrapper for all Meta messages
type MetaMessage struct {
	Type    string      `yaml:"type"`
	Version int         `yaml:"version"`
	Payload interface{} `yaml:"payload"`
}

// PlanTaskResponse is the expected payload for "plan_task"
type PlanTaskResponse struct {
	TaskID             string                `yaml:"task_id"`
	AcceptanceCriteria []AcceptanceCriterion `yaml:"acceptance_criteria"`
}

type AcceptanceCriterion struct {
	ID          string `yaml:"id"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
	Critical    bool   `yaml:"critical"`
	Passed      bool   `yaml:"passed"` // Added for context summary
}

// NextActionResponse is the expected payload for "next_action"
type NextActionResponse struct {
	Decision   Decision   `yaml:"decision"`
	WorkerCall WorkerCall `yaml:"worker_call,omitempty"`
}

type Decision struct {
	Action string `yaml:"action"` // "run_worker" | "mark_complete" | "ask_human" | "abort"
	Reason string `yaml:"reason"`
}

type WorkerCall struct {
	WorkerType string `yaml:"worker_type"`
	Mode       string `yaml:"mode"`
	Prompt     string `yaml:"prompt"`
}

// CompletionAssessmentResponse is the expected payload for "completion_assessment"
type CompletionAssessmentResponse struct {
	AllCriteriaSatisfied bool              `yaml:"all_criteria_satisfied"`
	Summary              string            `yaml:"summary"`
	ByCriterion          []CriterionResult `yaml:"by_criterion"`
}

type CriterionResult struct {
	ID      string `yaml:"id"`
	Status  string `yaml:"status"` // "passed" | "failed"
	Comment string `yaml:"comment"`
}

// TaskSummary is a simplified view of the task for the Meta agent
type TaskSummary struct {
	Title              string
	State              string
	AcceptanceCriteria []AcceptanceCriterion
	WorkerRunsCount    int
}
