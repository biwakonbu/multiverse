package agenttools

import (
	"context"
	"fmt"
	"strconv"
)

// CodexProvider builds ExecPlan for Codex CLI.
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
		DefaultModel:  p.model,
		SupportsStdin: true,
		Notes:         "Supports exec/chat; model/temperature/max-tokens flags are passed through.",
	}
}

func (p *CodexProvider) Build(_ context.Context, req Request) (ExecPlan, error) {
	if err := ensurePrompt(req.Prompt); err != nil {
		return ExecPlan{}, err
	}

	mode := req.Mode
	if mode == "" {
		mode = "exec"
	}

	args := []string{}

	switch mode {
	case "exec":
		args = append(args, "exec")

		// Sandbox scope (default: workspace-write)
		sandbox := "workspace-write"
		if v, ok := req.ToolSpecific["sandbox"].(string); ok && v != "" {
			sandbox = v
		}
		args = append(args, "--sandbox", sandbox)

		// JSON output for machine readability
		args = append(args, "--json")

		// Working directory inside container (fallback to /workspace/project)
		cwd := req.Workdir
		if cwd == "" {
			cwd = "/workspace/project"
		}
		args = append(args, "--cwd", cwd)

	case "chat":
		args = append(args, "chat")

	default:
		return ExecPlan{}, fmt.Errorf("%w: %s", ErrUnsupportedMode, mode)
	}

	// Model selection (request overrides config)
	if model := nonEmpty(req.Model, p.model); model != "" {
		args = append(args, "--model", model)
	}

	// Sampling controls
	if req.Temperature != nil {
		args = append(args, "--temperature", strconv.FormatFloat(*req.Temperature, 'f', 2, 64))
	}
	if req.MaxTokens != nil {
		args = append(args, "--max-tokens", strconv.Itoa(*req.MaxTokens))
	}

	// Provider defaults first, then request flags
	args = append(args, p.flags...)
	args = append(args, req.Flags...)

	plan := ExecPlan{
		Command: p.cliPath,
		Args:    args,
		Env:     mergeEnv(p.env, req.ExtraEnv),
		Workdir: req.Workdir,
		Timeout: req.Timeout,
	}

	if req.UseStdin {
		plan.Args = append(plan.Args, "--stdin")
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
