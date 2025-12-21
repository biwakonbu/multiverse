package agenttools

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// DefaultCodexModel は Codex CLI のデフォルトモデル（Worker 用）
// NOTE: モデル ID は OpenAI/Codex CLI の仕様に依存するため、必要に応じて LLMConfig で上書き可能。
const DefaultCodexModel = "gpt-5.2-codex"

// DefaultMetaModel は Meta-agent 用のデフォルトモデル
const DefaultMetaModel = "gpt-5.2"

// DefaultReasoningEffort は思考の深さのデフォルト値
const DefaultReasoningEffort = "medium"

// CodexProvider builds ExecPlan for Codex CLI.
// Codex CLI 0.65.0 対応。Docker コンテナ内での実行を前提とする。
type CodexProvider struct {
	cliPath string
	model   string
	env     map[string]string
	flags   []string
}

// NewCodexProvider constructs a CodexProvider from config.
func NewCodexProvider(cfg ProviderConfig) *CodexProvider {
	return &CodexProvider{
		cliPath: nonEmpty(cfg.CLIPath, "codex"),
		model:   cfg.Model,
		env:     mergeEnv(nil, cfg.ExtraEnv),
		flags:   append([]string{}, cfg.Flags...),
	}
}

func (p *CodexProvider) Kind() string {
	return "codex-cli"
}

func (p *CodexProvider) Capabilities() Capability {
	return Capability{
		Kind:          p.Kind(),
		DefaultModel:  nonEmpty(p.model, DefaultCodexModel),
		SupportsStdin: true,
		Notes:         "Codex CLI 0.65.0. Docker 内実行専用。exec モードのみサポート。",
	}
}

// Build は Codex CLI の実行計画を生成する。
// Codex CLI 0.65.0 の仕様に準拠:
//   - サンドボックス・承認を無効化（Docker が外部サンドボックスとして機能）
//   - 作業ディレクトリは -C フラグで指定
//   - 設定オーバーライドは -c フラグで指定（TOML 形式）
//   - stdin 入力は PROMPT に "-" を指定
//
// ToolSpecific オプション:
//   - docker_mode: bool - true の場合、Docker 内実行用フラグを追加（デフォルト: true）
//   - json_output: bool - true の場合、--json フラグを追加（デフォルト: true）
func (p *CodexProvider) Build(_ context.Context, req Request) (ExecPlan, error) {
	if err := ensurePrompt(req.Prompt); err != nil {
		return ExecPlan{}, err
	}

	// exec モードのみサポート（chat サブコマンドは Codex CLI に存在しない）
	mode := req.Mode
	if mode == "" {
		mode = "exec"
	}
	if mode != "exec" {
		return ExecPlan{}, fmt.Errorf("%w: %s (only 'exec' is supported)", ErrUnsupportedMode, mode)
	}

	// Docker モードかどうか（デフォルト: true = Worker 実行用）
	dockerMode := true
	if v, ok := req.ToolSpecific["docker_mode"].(bool); ok {
		dockerMode = v
	}

	// JSON 出力かどうか（デフォルト: true）
	jsonOutput := true
	if v, ok := req.ToolSpecific["json_output"].(bool); ok {
		jsonOutput = v
	}

	args := []string{"exec"}

	// Docker 内実行時のみ: サンドボックス・承認を無効化
	// 参照: docs/design/sandbox-policy.md
	if dockerMode {
		args = append(args, "--dangerously-bypass-approvals-and-sandbox")
	}

	// 作業ディレクトリ（-C フラグ）
	// Docker モード時はデフォルト /workspace/project、それ以外は指定がある場合のみ
	if req.Workdir != "" {
		args = append(args, "-C", req.Workdir)
	} else if dockerMode {
		args = append(args, "-C", "/workspace/project")
	}

	// JSON 出力（機械可読形式）
	if jsonOutput {
		args = append(args, "--json")
	}

	// モデル指定（デフォルト: gpt-5.2-codex）
	model := nonEmpty(req.Model, p.model, DefaultCodexModel)
	model = ResolveOpenAIModelID(model)
	args = append(args, "-m", model)

	// 思考の深さ（デフォルト: medium）
	// Request.ReasoningEffort または ToolSpecific から取得
	reasoningEffort := DefaultReasoningEffort
	if req.ReasoningEffort != "" {
		reasoningEffort = req.ReasoningEffort
	} else if v, ok := req.ToolSpecific["reasoning_effort"].(string); ok && v != "" {
		reasoningEffort = v
	}

	// Codex CLI / OpenAI API が受け付けない値が来た場合に安全側へ丸める。
	// 例: "xhigh" など未サポート値は "high" にフォールバックする。
	normalizedEffort := strings.ToLower(strings.TrimSpace(reasoningEffort))
	switch normalizedEffort {
	case "xhigh", "extra_high", "very_high":
		normalizedEffort = "high"
	case "":
		normalizedEffort = DefaultReasoningEffort
	}
	if normalizedEffort != "none" && normalizedEffort != "low" && normalizedEffort != "medium" && normalizedEffort != "high" {
		normalizedEffort = DefaultReasoningEffort
	}
	reasoningEffort = normalizedEffort
	args = append(args, "-c", fmt.Sprintf("reasoning_effort=%s", reasoningEffort))

	// 追加の config オーバーライド（-c フラグで TOML 形式）
	if req.Temperature != nil {
		args = append(args, "-c", fmt.Sprintf("temperature=%s", strconv.FormatFloat(*req.Temperature, 'f', 2, 64)))
	}
	if req.MaxTokens != nil {
		args = append(args, "-c", fmt.Sprintf("max_tokens=%d", *req.MaxTokens))
	}

	// 追加フラグ（Provider デフォルト → Request）
	args = append(args, p.flags...)
	args = append(args, req.Flags...)

	plan := ExecPlan{
		Command: p.cliPath,
		Args:    args,
		Env:     mergeEnv(p.env, req.ExtraEnv),
		Workdir: req.Workdir,
		Timeout: req.Timeout,
	}

	// プロンプト（stdin 使用時は "-" を指定）
	if req.UseStdin {
		plan.Args = append(plan.Args, "-")
		plan.Stdin = req.Prompt
	} else {
		// Append prompt as positional argument
		plan.Args = append(plan.Args, req.Prompt)
	}

	return plan, nil
}

func init() {
	Register("codex-cli", func(cfg ProviderConfig) (AgentToolProvider, error) {
		return NewCodexProvider(cfg), nil
	})
}
