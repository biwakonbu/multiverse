package meta

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"github.com/biwakonbu/agent-runner/internal/agenttools"
	"github.com/biwakonbu/agent-runner/internal/logging"
	"gopkg.in/yaml.v3"
)

// CLIProvider は CLI ベースの LLM プロバイダを抽象化するインターフェース
type CLIProvider interface {
	Name() string
	Decompose(ctx context.Context, req *DecomposeRequest) (*DecomposeResponse, error)
	PlanTask(ctx context.Context, prdText string) (*PlanTaskResponse, error)
	NextAction(ctx context.Context, taskSummary *TaskSummary) (*NextActionResponse, error)
	CompletionAssessment(ctx context.Context, taskSummary *TaskSummary) (*CompletionAssessmentResponse, error)
	TestConnection(ctx context.Context) error
}

// DefaultMetaAgentTimeout は Meta-agent 用のデフォルトタイムアウト（10分）
// Codex CLI での LLM 処理は時間がかかるため、十分な時間を確保する
const DefaultMetaAgentTimeout = 10 * time.Minute

// CodexCLIProvider は Codex CLI を使用するプロバイダ
type CodexCLIProvider struct {
	model        string
	systemPrompt string
	logger       *slog.Logger
}

// NewCodexCLIProvider は新しい Codex CLI プロバイダを作成する
func NewCodexCLIProvider(model, systemPrompt string) *CodexCLIProvider {
	return &CodexCLIProvider{
		model:        model,
		systemPrompt: systemPrompt,
		logger:       logging.WithComponent(slog.Default(), "meta-codex-cli"),
	}
}

// SetLogger はカスタムロガーを設定する
func (p *CodexCLIProvider) SetLogger(logger *slog.Logger) {
	p.logger = logging.WithComponent(logger, "meta-codex-cli")
}

// Name はプロバイダ名を返す
func (p *CodexCLIProvider) Name() string {
	return "codex-cli"
}

// TestConnection は Codex CLI セッションの存在を検証する
func (p *CodexCLIProvider) TestConnection(ctx context.Context) error {
	logger := logging.WithTraceID(p.logger, ctx)

	// codex --version または codex auth status でセッション確認
	cmd := exec.CommandContext(ctx, "codex", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("codex CLI not found or not authenticated",
			slog.String("output", string(output)),
			slog.Any("error", err),
		)
		return fmt.Errorf("codex CLI セッションが見つかりません: %w (出力: %s)", err, string(output))
	}

	logger.Info("codex CLI session verified", slog.String("version_output", strings.TrimSpace(string(output))))
	return nil
}

// callCodexExec は codex exec コマンドを実行し、YAML 応答を取得する
// agenttools パッケージを使用して共通のフラグ構築ロジックを適用する。
func (p *CodexCLIProvider) callCodexExec(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	logger := logging.WithTraceID(p.logger, ctx)
	start := time.Now()

	// システムプロンプトとユーザープロンプトを結合
	fullPrompt := systemPrompt + "\n\n" + userPrompt

	// agenttools を使用して ExecPlan を生成
	model := p.model
	if model == "" {
		model = agenttools.DefaultMetaModel // Meta-agent 用デフォルト
	}

	req := agenttools.Request{
		Prompt:          fullPrompt,
		Model:           model,
		ReasoningEffort: agenttools.DefaultReasoningEffort,
		Timeout:         DefaultMetaAgentTimeout, // LLM 処理に十分な時間を確保
		UseStdin:        true,                    // 長いプロンプトは stdin で渡す
		ToolSpecific: map[string]interface{}{
			"docker_mode": false, // ホスト上で直接実行
			"json_output": false, // Meta-agent は YAML 出力を期待
		},
	}

	provider := agenttools.NewCodexProvider(agenttools.ProviderConfig{
		Kind:  "codex-cli",
		Model: model,
	})

	plan, err := provider.Build(ctx, req)
	if err != nil {
		logger.Error("failed to build exec plan", slog.Any("error", err))
		return "", fmt.Errorf("ExecPlan 構築失敗: %w", err)
	}

	logger.Info("calling codex CLI",
		slog.Int("prompt_length", len(fullPrompt)),
		slog.String("model", model),
	)
	logger.Debug("codex exec plan",
		slog.String("command", plan.Command),
		slog.Any("args", plan.Args),
	)

	// agenttools.Execute を使用してコマンドを実行
	result := agenttools.Execute(ctx, plan)
	if result.Error != nil {
		logger.Error("codex CLI call failed",
			slog.String("output", result.Output),
			slog.Int("exit_code", result.ExitCode),
			slog.Any("error", result.Error),
			logging.LogDuration(start),
		)
		return "", fmt.Errorf("codex CLI 呼び出し失敗: %w (出力: %s)", result.Error, result.Output)
	}

	response := strings.TrimSpace(result.Output)
	logger.Info("codex CLI call completed",
		slog.Int("response_length", len(response)),
		logging.LogDuration(start),
	)
	logger.Debug("codex CLI response", slog.String("response", response))

	return response, nil
}

// Decompose はユーザー入力からタスクを分解する
func (p *CodexCLIProvider) Decompose(ctx context.Context, req *DecomposeRequest) (*DecomposeResponse, error) {
	logger := logging.WithTraceID(p.logger, ctx)

	systemPrompt := decomposeSystemPrompt
	if p.systemPrompt != "" {
		systemPrompt = p.systemPrompt
	}
	userPrompt := buildDecomposeUserPrompt(req)

	logger.Info("calling codex CLI for decompose",
		slog.String("user_input_length", fmt.Sprintf("%d", len(req.UserInput))),
		slog.Int("existing_tasks", len(req.Context.ExistingTasks)),
	)

	resp, err := p.callCodexExec(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("codex CLI call failed: %w", err)
	}

	// YAML を抽出してパース
	resp = extractYAML(resp)

	var msg MetaMessage
	if err := yaml.Unmarshal([]byte(resp), &msg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w\nResponse: %s", err, resp)
	}

	payloadBytes, err := yaml.Marshal(msg.Payload)
	if err != nil {
		return nil, err
	}

	var decompose DecomposeResponse
	if err := yaml.Unmarshal(payloadBytes, &decompose); err != nil {
		return nil, fmt.Errorf("failed to parse decompose response: %w", err)
	}

	logger.Info("decompose completed",
		slog.Int("phases", len(decompose.Phases)),
		slog.Int("potential_conflicts", len(decompose.PotentialConflicts)),
	)

	return &decompose, nil
}

// PlanTask は PRD から受け入れ条件を生成する
func (p *CodexCLIProvider) PlanTask(ctx context.Context, prdText string) (*PlanTaskResponse, error) {
	systemPrompt := `You are a Meta-agent that plans software development tasks.
Your goal is to read a PRD and break it down into Acceptance Criteria.
Output MUST be a YAML block with the following structure:
type: plan_task
version: 1
payload:
  task_id: "TASK-..."
  acceptance_criteria:
    - id: "AC-1"
      description: "..."
      type: "e2e"
      critical: true
`
	if p.systemPrompt != "" {
		systemPrompt = p.systemPrompt
	}
	userPrompt := fmt.Sprintf("PRD:\n%s\n\nGenerate the plan.", prdText)

	resp, err := p.callCodexExec(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	// Extract YAML from response
	resp = extractYAML(resp)

	var msg MetaMessage
	if err := yaml.Unmarshal([]byte(resp), &msg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w\nResponse: %s", err, resp)
	}

	payloadBytes, err := yaml.Marshal(msg.Payload)
	if err != nil {
		return nil, err
	}
	var plan PlanTaskResponse
	if err := yaml.Unmarshal(payloadBytes, &plan); err != nil {
		return nil, err
	}

	return &plan, nil
}

// NextAction は次のアクションを決定する
func (p *CodexCLIProvider) NextAction(ctx context.Context, taskSummary *TaskSummary) (*NextActionResponse, error) {
	systemPrompt := `You are a Meta-agent that orchestrates a coding task.
Decide the next action based on the current context.
Output MUST be a YAML block with type: next_action.
`
	if p.systemPrompt != "" {
		systemPrompt = p.systemPrompt
	}
	contextSummary := fmt.Sprintf("Task: %s\nState: %s\nACs: %v\nWorkerRuns: %d",
		taskSummary.Title, taskSummary.State, len(taskSummary.AcceptanceCriteria), taskSummary.WorkerRunsCount)

	userPrompt := fmt.Sprintf("Context:\n%s\n\nDecide next action.", contextSummary)

	resp, err := p.callCodexExec(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	// Extract YAML from response
	resp = extractYAML(resp)

	var msg MetaMessage
	if err := yaml.Unmarshal([]byte(resp), &msg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w\nResponse: %s", err, resp)
	}

	payloadBytes, err := yaml.Marshal(msg.Payload)
	if err != nil {
		return nil, err
	}
	var action NextActionResponse
	if err := yaml.Unmarshal(payloadBytes, &action); err != nil {
		return nil, err
	}

	return &action, nil
}

// CompletionAssessment はタスク完了を評価する
func (p *CodexCLIProvider) CompletionAssessment(ctx context.Context, taskSummary *TaskSummary) (*CompletionAssessmentResponse, error) {
	systemPrompt := `You are a Meta-agent evaluating task completion.
Review the Acceptance Criteria and Worker execution results.
Output MUST be a YAML block with type: completion_assessment.

Example format:
type: completion_assessment
version: 1
payload:
  all_criteria_satisfied: true
  summary: "All acceptance criteria met"
  by_criterion:
    - id: "AC-1"
      status: "passed"
      comment: "Feature X successfully implemented"
`
	if p.systemPrompt != "" {
		systemPrompt = p.systemPrompt
	}

	// Format acceptance criteria for LLM
	acText := ""
	for _, ac := range taskSummary.AcceptanceCriteria {
		acText += fmt.Sprintf("- %s: %s\n", ac.ID, ac.Description)
	}

	// Format worker runs for LLM
	workerText := ""
	for _, run := range taskSummary.WorkerRuns {
		workerText += fmt.Sprintf("- Run %s: exit_code=%d, summary=%s\n", run.ID, run.ExitCode, run.Summary)
	}

	userPrompt := fmt.Sprintf(`Task: %s
State: %s

Acceptance Criteria:
%s

Worker Execution Results:
%s

Evaluate whether all acceptance criteria are satisfied.`,
		taskSummary.Title, taskSummary.State, acText, workerText)

	resp, err := p.callCodexExec(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	// Extract YAML from response
	resp = extractYAML(resp)

	var msg MetaMessage
	if err := yaml.Unmarshal([]byte(resp), &msg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w\nResponse: %s", err, resp)
	}

	payloadBytes, err := yaml.Marshal(msg.Payload)
	if err != nil {
		return nil, err
	}
	var assessment CompletionAssessmentResponse
	if err := yaml.Unmarshal(payloadBytes, &assessment); err != nil {
		return nil, err
	}

	return &assessment, nil
}
