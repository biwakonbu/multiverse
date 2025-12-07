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
	WorkerType   string                 `yaml:"worker_type"`
	Mode         string                 `yaml:"mode"`
	Prompt       string                 `yaml:"prompt"`
	Model        string                 `yaml:"model,omitempty"`
	Temperature  *float64               `yaml:"temperature,omitempty"`
	MaxTokens    *int                   `yaml:"max_tokens,omitempty"`
	CLIPath      string                 `yaml:"cli_path,omitempty"`
	Flags        []string               `yaml:"flags,omitempty"`
	Env          map[string]string      `yaml:"env,omitempty"`
	ToolSpecific map[string]interface{} `yaml:"tool_specific,omitempty"`
	Workdir      string                 `yaml:"workdir,omitempty"`
	UseStdin     bool                   `yaml:"use_stdin,omitempty"`
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

// WorkerRunSummary is a summary of a single worker run
type WorkerRunSummary struct {
	ID       string `yaml:"id"`
	ExitCode int    `yaml:"exit_code"`
	Summary  string `yaml:"summary"`
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
	UserInput string           `yaml:"user_input"` // ユーザーの入力メッセージ
	Context   DecomposeContext `yaml:"context"`    // コンテキスト情報
}

// DecomposeContext はタスク分解時のコンテキスト情報
type DecomposeContext struct {
	WorkspacePath       string                `yaml:"workspace_path"`       // プロジェクトパス
	ExistingTasks       []ExistingTaskSummary `yaml:"existing_tasks"`       // 既存タスク一覧
	ConversationHistory []ConversationMessage `yaml:"conversation_history"` // 会話履歴
}

// ExistingTaskSummary は既存タスクの要約（分解時の参照用）
type ExistingTaskSummary struct {
	ID           string   `yaml:"id"`
	Title        string   `yaml:"title"`
	Status       string   `yaml:"status"`
	Dependencies []string `yaml:"dependencies,omitempty"`
	PhaseName    string   `yaml:"phase_name,omitempty"`
}

// ConversationMessage はチャット履歴の1メッセージ
type ConversationMessage struct {
	Role    string `yaml:"role"` // user | assistant | system
	Content string `yaml:"content"`
}

// DecomposeResponse はタスク分解結果
type DecomposeResponse struct {
	Understanding      string              `yaml:"understanding"`       // ユーザー意図の理解
	Phases             []DecomposedPhase   `yaml:"phases"`              // フェーズ別タスク
	PotentialConflicts []PotentialConflict `yaml:"potential_conflicts"` // 潜在的なコンフリクト
}

// DecomposedPhase は分解されたフェーズ（概念設計/実装設計/実装）
type DecomposedPhase struct {
	Name      string           `yaml:"name"`      // フェーズ名
	Milestone string           `yaml:"milestone"` // マイルストーン名
	Tasks     []DecomposedTask `yaml:"tasks"`     // フェーズ内タスク
}

// DecomposedTask は分解された個別タスク
type DecomposedTask struct {
	ID                 string   `yaml:"id"`                  // 一時ID（保存時に正式IDに置換）
	Title              string   `yaml:"title"`               // タスクタイトル
	Description        string   `yaml:"description"`         // 詳細説明
	AcceptanceCriteria []string `yaml:"acceptance_criteria"` // 達成条件
	Dependencies       []string `yaml:"dependencies"`        // 依存タスクID（同バッチ内の一時ID参照可）
	WBSLevel           int      `yaml:"wbs_level"`           // WBS階層レベル
	EstimatedEffort    string   `yaml:"estimated_effort"`    // 推定工数（small/medium/large）
}

// PotentialConflict はファイルコンフリクトの可能性
type PotentialConflict struct {
	File    string   `yaml:"file"`    // 対象ファイル
	Tasks   []string `yaml:"tasks"`   // 関連タスクID
	Warning string   `yaml:"warning"` // 警告メッセージ
}
