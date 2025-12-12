package persistence

import (
	"time"
)

// --- Design Models ---

type WBS struct {
	WBSID       string      `json:"wbs_id"`
	ProjectRoot string      `json:"project_root"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	RootNodeID  string      `json:"root_node_id"`
	NodeIndex   []NodeIndex `json:"node_index"`
}

type NodeIndex struct {
	NodeID   string   `json:"node_id"`
	ParentID *string  `json:"parent_id"` // Nullable
	Children []string `json:"children"`
}

type NodeDesign struct {
	NodeID             string        `json:"node_id"`
	WBSID              string        `json:"wbs_id"`
	Name               string        `json:"name"`
	Summary            string        `json:"summary"`
	Kind               string        `json:"kind"` // feature, refactor, bugfix, ...
	Priority           string        `json:"priority"`
	Estimate           Estimate      `json:"estimate"`
	Dependencies       []string      `json:"dependencies"`
	AcceptanceCriteria []string      `json:"acceptance_criteria"`
	DesignNotes        []string      `json:"design_notes"`
	SuggestedImpl      SuggestedImpl `json:"suggested_impl"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
	CreatedBy          string        `json:"created_by"`
}

type Estimate struct {
	StoryPoints int    `json:"story_points"`
	Difficulty  string `json:"difficulty"`
}

type SuggestedImpl struct {
	Language    string   `json:"language"`
	Framework   string   `json:"framework"`
	ModulePaths []string `json:"module_paths"`
	FilePaths   []string `json:"file_paths"`  // Added for compatibility
	Constraints []string `json:"constraints"` // Added for compatibility
}

// --- State Models ---

type NodesRuntime struct {
	Nodes []NodeRuntime `json:"nodes"`
}

type NodeRuntime struct {
	NodeID         string             `json:"node_id"`
	Status         string             `json:"status"` // planned, in_progress, implemented, verified, blocked, obsolete
	Implementation NodeImplementation `json:"implementation"`
	Verification   NodeVerification   `json:"verification"`
	Notes          []NodeNote         `json:"notes"`
}

type NodeImplementation struct {
	Files          []string  `json:"files"`
	LastModifiedAt time.Time `json:"last_modified_at"`
	LastModifiedBy string    `json:"last_modified_by"`
}

type NodeVerification struct {
	Status         string    `json:"status"` // not_tested, passed, failed, flaky
	LastTestTaskID string    `json:"last_test_task_id"`
	LastTestAt     time.Time `json:"last_test_at"`
}

type NodeNote struct {
	At   time.Time `json:"at"`
	By   string    `json:"by"`
	Text string    `json:"text"`
}

type TasksState struct {
	Tasks     []TaskState `json:"tasks"`
	QueueMeta QueueMeta   `json:"queue_meta"`
}

type TaskState struct {
	TaskID        string                 `json:"task_id"`
	NodeID        string                 `json:"node_id"`
	Kind          string                 `json:"kind"`   // planning, implementation, test, refactor, analysis
	Status        string                 `json:"status"` // pending, running, succeeded, failed, canceled, skipped
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	ScheduledBy   string                 `json:"scheduled_by"`
	AssignedAgent string                 `json:"assigned_agent"`
	Priority      int                    `json:"priority"`
	Inputs        map[string]interface{} `json:"inputs"` // Flexible inputs (goal, constraints, etc.)
	Outputs       TaskOutputs            `json:"outputs"`
}

type TaskOutputs struct {
	Status    string                 `json:"status"`
	Artifacts map[string]interface{} `json:"artifacts"` // Flexible artifacts
}

type QueueMeta struct {
	LastScheduledAt time.Time `json:"last_scheduled_at"`
	NextTaskIDSeq   int64     `json:"next_task_id_seq"`
}

type AgentsState struct {
	Agents []AgentState `json:"agents"`
}

type AgentState struct {
	AgentID      string   `json:"agent_id"`
	Kind         string   `json:"kind"` // code, test, ...
	MaxParallel  int      `json:"max_parallel"`
	RunningTasks []string `json:"running_tasks"`
	Capabilities []string `json:"capabilities"`
}

// --- History Models ---

type Action struct {
	ID          string                 `json:"id"`
	At          time.Time              `json:"at"`
	Kind        string                 `json:"kind"`
	WorkspaceID string                 `json:"workspace_id"`
	Payload     map[string]interface{} `json:"payload"`
}
