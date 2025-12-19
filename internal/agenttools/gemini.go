package agenttools

import (
	"context"
	"fmt"
	"strconv"
)

// DefaultGeminiModel defines the default model for Gemini CLI.
// 公式ドキュメント: https://ai.google.dev/gemini-api/docs/models
// CLI フラグ:
//   - -m, --model <model>: モデル指定
//   - -p, --headless-mode: 非対話モード
//   - --output-format <format>: 出力形式 (json, text, stream-json)
//   - --yolo: ツール呼び出しを自動承認
//   - --sandbox: Docker コンテナ内で実行
const DefaultGeminiModel = "gemini-3-flash-preview"

// GeminiProvider builds ExecPlan for Gemini CLI.
// Assumes the use of Google's open-source Gemini CLI or compatible interface.
type GeminiProvider struct {
	cliPath string
	model   string
	env     map[string]string
	flags   []string
}

// NewGeminiProvider constructs a GeminiProvider from config.
func NewGeminiProvider(cfg ProviderConfig) *GeminiProvider {
	return &GeminiProvider{
		cliPath: nonEmpty(cfg.CLIPath, "gemini"),
		model:   cfg.Model,
		env:     mergeEnv(nil, cfg.ExtraEnv),
		flags:   append([]string{}, cfg.Flags...),
	}
}

func (p *GeminiProvider) Kind() string {
	return "gemini-cli"
}

func (p *GeminiProvider) Capabilities() Capability {
	return Capability{
		Kind:          p.Kind(),
		DefaultModel:  nonEmpty(p.model, DefaultGeminiModel),
		SupportsStdin: true,
		Notes:         "Generic Gemini CLI wrapper. Assumes `gemini -p [prompt] --model [model]` interface.",
	}
}

// Build generates the execution plan for Gemini CLI.
func (p *GeminiProvider) Build(_ context.Context, req Request) (ExecPlan, error) {
	if err := ensurePrompt(req.Prompt); err != nil {
		return ExecPlan{}, err
	}

	// Gemini CLI usually just takes the prompt as an argument or stdin.
	mode := req.Mode
	if mode == "" {
		mode = "exec"
	}
	// For now, only 'exec' mode is supported in this wrapper (running a command/prompt).
	// If specific subcommands like 'chat' are verifiable, they can be added.
	if mode != "exec" {
		return ExecPlan{}, fmt.Errorf("%w: %s (only 'exec' is supported)", ErrUnsupportedMode, mode)
	}

	args := []string{}

	// JSON output support check
	jsonOutput := true
	if v, ok := req.ToolSpecific["json_output"].(bool); ok {
		jsonOutput = v
	}

	// Auto-accept tool calls (non-interactive execution)
	autoAccept := true
	if v, ok := req.ToolSpecific["auto_accept"].(bool); ok {
		autoAccept = v
	}
	if v, ok := req.ToolSpecific["yolo"].(bool); ok {
		autoAccept = v
	}

	// Model specification
	model := nonEmpty(req.Model, p.model, DefaultGeminiModel)
	args = append(args, "--model", model)

	// Temperature mapping
	if req.Temperature != nil {
		args = append(args, "--temperature", strconv.FormatFloat(*req.Temperature, 'f', 2, 64))
	}

	// MaxTokens mapping
	if req.MaxTokens != nil {
		args = append(args, "--max-output-tokens", strconv.Itoa(*req.MaxTokens))
	}

	if jsonOutput {
		args = append(args, "--output-format", "json")
	}
	if autoAccept {
		args = append(args, "--yolo")
	}

	// Extra flags
	args = append(args, p.flags...)
	args = append(args, req.Flags...)

	plan := ExecPlan{
		Command: p.cliPath,
		Args:    args,
		Env:     mergeEnv(p.env, req.ExtraEnv),
		Workdir: req.Workdir,
		Timeout: req.Timeout,
	}

	// Prompt handling
	if req.UseStdin {
		plan.Args = append(plan.Args, "-p", "-")
		plan.Stdin = req.Prompt
	} else {
		plan.Args = append(plan.Args, "-p", req.Prompt)
	}

	return plan, nil
}

func init() {
	Register("gemini-cli", func(cfg ProviderConfig) (AgentToolProvider, error) {
		return NewGeminiProvider(cfg), nil
	})
}
