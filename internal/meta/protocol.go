package meta

// Protocol definitions for Meta-agent communication

// Common wrapper for all Meta messages
type MetaMessage struct {
	Type    string      `yaml:"type" json:"type"`
	Version int         `yaml:"version" json:"version"`
	Payload interface{} `yaml:"payload" json:"payload"`
}

// PlanTaskResponse is the expected payload for "plan_task"
type PlanTaskResponse struct {
	TaskID             string                `yaml:"task_id" json:"task_id"`
	AcceptanceCriteria []AcceptanceCriterion `yaml:"acceptance_criteria" json:"acceptance_criteria"`
}

type AcceptanceCriterion struct {
	ID          string `yaml:"id" json:"id"`
	Description string `yaml:"description" json:"description"`
	Type        string `yaml:"type" json:"type"`
	Critical    bool   `yaml:"critical" json:"critical"`
	Passed      bool   `yaml:"passed" json:"passed"` // Added for context summary
}

// NextActionResponse is the expected payload for "next_action"
type NextActionResponse struct {
	Decision   Decision   `yaml:"decision" json:"decision"`
	WorkerCall WorkerCall `yaml:"worker_call,omitempty" json:"worker_call,omitempty"`
}

type Decision struct {
	Action string `yaml:"action" json:"action"` // "run_worker" | "mark_complete" | "ask_human" | "abort"
	Reason string `yaml:"reason" json:"reason"`
}

type WorkerCall struct {
	WorkerType      string                 `yaml:"worker_type" json:"worker_type"`
	Mode            string                 `yaml:"mode" json:"mode"`
	Prompt          string                 `yaml:"prompt" json:"prompt"`
	Model           string                 `yaml:"model,omitempty" json:"model,omitempty"`
	Temperature     *float64               `yaml:"temperature,omitempty" json:"temperature,omitempty"`
	MaxTokens       *int                   `yaml:"max_tokens,omitempty" json:"max_tokens,omitempty"`
	ReasoningEffort string                 `yaml:"reasoning_effort,omitempty" json:"reasoning_effort,omitempty"` // "low" | "medium" | "high"
	CLIPath         string                 `yaml:"cli_path,omitempty" json:"cli_path,omitempty"`
	Flags           []string               `yaml:"flags,omitempty" json:"flags,omitempty"`
	Env             map[string]string      `yaml:"env,omitempty" json:"env,omitempty"`
	ToolSpecific    map[string]interface{} `yaml:"tool_specific,omitempty" json:"tool_specific,omitempty"`
	Workdir         string                 `yaml:"workdir,omitempty" json:"workdir,omitempty"`
	UseStdin        bool                   `yaml:"use_stdin,omitempty" json:"use_stdin,omitempty"`
}

// CompletionAssessmentResponse is the expected payload for "completion_assessment"
type CompletionAssessmentResponse struct {
	AllCriteriaSatisfied bool              `yaml:"all_criteria_satisfied" json:"all_criteria_satisfied"`
	Summary              string            `yaml:"summary" json:"summary"`
	ByCriterion          []CriterionResult `yaml:"by_criterion" json:"by_criterion"`
}

type CriterionResult struct {
	ID      string `yaml:"id" json:"id"`
	Status  string `yaml:"status" json:"status"` // "passed" | "failed"
	Comment string `yaml:"comment" json:"comment"`
}

// WorkerRunSummary is a summary of a single worker run
type WorkerRunSummary struct {
	ID       string `yaml:"id" json:"id"`
	ExitCode int    `yaml:"exit_code" json:"exit_code"`
	Summary  string `yaml:"summary" json:"summary"`
}

// TaskSummary is a simplified view of the task for the Meta agent
type TaskSummary struct {
	Title              string
	State              string
	AcceptanceCriteria []AcceptanceCriterion
	WorkerRunsCount    int
	WorkerRuns         []WorkerRunSummary
}

// ============================================================================
// Decompose Protocol (v2.0): チャットからタスク分解
// ============================================================================

// DecomposeRequest はチャット入力からタスク分解を要求するリクエスト
type DecomposeRequest struct {
	UserInput string           `yaml:"user_input" json:"user_input"` // ユーザーの入力メッセージ
	Context   DecomposeContext `yaml:"context" json:"context"`       // コンテキスト情報
}

// DecomposeContext はタスク分解時のコンテキスト情報
type DecomposeContext struct {
	WorkspacePath       string                `yaml:"workspace_path" json:"workspace_path"`             // プロジェクトパス
	ExistingTasks       []ExistingTaskSummary `yaml:"existing_tasks" json:"existing_tasks"`             // 既存タスク一覧
	ConversationHistory []ConversationMessage `yaml:"conversation_history" json:"conversation_history"` // 会話履歴
}

// ExistingTaskSummary は既存タスクの要約（分解時の参照用）
type ExistingTaskSummary struct {
	ID           string   `yaml:"id" json:"id"`
	Title        string   `yaml:"title" json:"title"`
	Status       string   `yaml:"status" json:"status"`
	Dependencies []string `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	PhaseName    string   `yaml:"phase_name,omitempty" json:"phase_name,omitempty"`
}

// ConversationMessage はチャット履歴の1メッセージ
type ConversationMessage struct {
	Role    string `yaml:"role" json:"role"` // user | assistant | system
	Content string `yaml:"content" json:"content"`
}

// DecomposeResponse はタスク分解結果
type DecomposeResponse struct {
	Understanding      string              `yaml:"understanding" json:"understanding"`             // ユーザー意図の理解
	Phases             []DecomposedPhase   `yaml:"phases" json:"phases"`                           // フェーズ別タスク
	PotentialConflicts []PotentialConflict `yaml:"potential_conflicts" json:"potential_conflicts"` // 潜在的なコンフリクト
}

// DecomposedPhase は分解されたフェーズ（概念設計/実装設計/実装）
type DecomposedPhase struct {
	Name      string           `yaml:"name" json:"name"`           // フェーズ名
	Milestone string           `yaml:"milestone" json:"milestone"` // マイルストーン名
	Tasks     []DecomposedTask `yaml:"tasks" json:"tasks"`         // フェーズ内タスク
}

// DecomposedTask は分解された個別タスク
type DecomposedTask struct {
	ID                 string   `yaml:"id" json:"id"`                                   // 一時ID（保存時に正式IDに置換）
	Title              string   `yaml:"title" json:"title"`                             // タスクタイトル
	Description        string   `yaml:"description" json:"description"`                 // 詳細説明
	AcceptanceCriteria []string `yaml:"acceptance_criteria" json:"acceptance_criteria"` // 達成条件
	Dependencies       []string `yaml:"dependencies" json:"dependencies"`               // 依存タスクID（同バッチ内の一時ID参照可）
	WBSLevel           int      `yaml:"wbs_level" json:"wbs_level"`                     // WBS階層レベル
	EstimatedEffort    string   `yaml:"estimated_effort" json:"estimated_effort"`       // 推定工数（small/medium/large）
}

// PotentialConflict はファイルコンフリクトの可能性
type PotentialConflict struct {
	File    string   `yaml:"file" json:"file"`       // 対象ファイル
	Tasks   []string `yaml:"tasks" json:"tasks"`     // 関連タスクID
	Warning string   `yaml:"warning" json:"warning"` // 警告メッセージ
}
