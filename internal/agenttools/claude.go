package agenttools

import (
	"context"
	"fmt"
)

// DefaultClaudeModel は Claude Code のデフォルトモデル。
// 参照: https://docs.anthropic.com/en/docs/claude-code
const DefaultClaudeModel = "claude-haiku-4-5-20251001"

// ClaudeProvider builds ExecPlan for Claude Code CLI.
// Wrapper for `claude-code` or `claude` CLI.
type ClaudeProvider struct {
	kind    string
	cliPath string
	model   string
	env     map[string]string
	flags   []string
}

// NewClaudeProvider constructs a ClaudeProvider from config.
func NewClaudeProvider(cfg ProviderConfig) *ClaudeProvider {
	return newClaudeProvider("claude-code", cfg)
}

func newClaudeProvider(kind string, cfg ProviderConfig) *ClaudeProvider {
	return &ClaudeProvider{
		kind:    kind,
		cliPath: nonEmpty(cfg.CLIPath, "claude"),
		model:   cfg.Model,
		env:     mergeEnv(nil, cfg.ExtraEnv),
		flags:   append([]string{}, cfg.Flags...),
	}
}

func (p *ClaudeProvider) Kind() string {
	return p.kind
}

func (p *ClaudeProvider) Capabilities() Capability {
	return Capability{
		Kind:          p.Kind(),
		DefaultModel:  nonEmpty(p.model, DefaultClaudeModel),
		SupportsStdin: true,
		Notes:         "Claude Code CLI wrapper. Assumes `claude -p [prompt]` interface.",
	}
}

// Build generates the execution plan for Claude Code CLI.
func (p *ClaudeProvider) Build(_ context.Context, req Request) (ExecPlan, error) {
	if err := ensurePrompt(req.Prompt); err != nil {
		return ExecPlan{}, err
	}

	mode := req.Mode
	if mode == "" {
		mode = "exec"
	}

	// claude-code usually handles conversation or single shot.
	// We map 'exec' to single shot or piped input.
	if mode != "exec" {
		return ExecPlan{}, fmt.Errorf("%w: %s (only 'exec' is supported)", ErrUnsupportedMode, mode)
	}

	// CLI実行フラグ構築
	// 重要: 非対話型で実行するために -p (--print) が必須
	args := []string{}

	// Model specification
	// デフォルトモデルも明示的に渡す（CLI側のデフォルトが変わっても影響を受けないように）
	model := nonEmpty(req.Model, p.model, DefaultClaudeModel)
	args = append(args, "--model", model)

	// Extra flags
	args = append(args, p.flags...)
	args = append(args, req.Flags...)

	plan := ExecPlan{
		Command: p.cliPath,
		Env:     mergeEnv(p.env, req.ExtraEnv),
		Workdir: req.Workdir,
		Timeout: req.Timeout,
	}

	// Prompt handling
	if req.UseStdin {
		// Piped prompt (stdin).
		// Conventions across providers: pass "-" as a placeholder and stream the prompt via stdin.
		args = append(args, "-p", "-")
		plan.Stdin = req.Prompt
	} else {
		// Normal argument execution
		args = append(args, "-p", req.Prompt)
	}

	plan.Args = args

	return plan, nil
}

func init() {
	Register("claude-code", func(cfg ProviderConfig) (AgentToolProvider, error) {
		return newClaudeProvider("claude-code", cfg), nil
	})
	// Backward compatible alias (docs/specifications and ISSUE.md used this name historically).
	Register("claude-code-cli", func(cfg ProviderConfig) (AgentToolProvider, error) {
		return newClaudeProvider("claude-code-cli", cfg), nil
	})
}
