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

// DefaultMetaAgentTimeout is 10 minutes for Meta processing
const DefaultMetaAgentTimeout = 10 * time.Minute

// CLIProvider は generic な LLM CLI（codex/claude/gemini）を agenttools 経由で扱う。
type CLIProvider struct {
	kind         string // "codex-cli", "claude-cli" or just "claude"
	cliPath      string // executable name or path
	model        string
	systemPrompt string
	flags        []string
	env          map[string]string
	toolSpecific map[string]interface{}
	logger       *slog.Logger
}

// Ensure CLIProvider implements Provider interface
var _ Provider = (*CLIProvider)(nil)

// CLIProviderOptions は CLI 実行の上書き設定を保持する。
type CLIProviderOptions struct {
	CLIPath      string
	Flags        []string
	Env          map[string]string
	ToolSpecific map[string]interface{}
}

// NewCLIProvider は CLI プロバイダを作成する。
func NewCLIProvider(kind, model, systemPrompt string) *CLIProvider {
	return NewCLIProviderWithOptions(kind, model, systemPrompt, CLIProviderOptions{})
}

// NewCLIProviderWithOptions は上書き設定付きの CLI プロバイダを作成する。
func NewCLIProviderWithOptions(kind, model, systemPrompt string, opts CLIProviderOptions) *CLIProvider {
	// Determine CLI path based on kind
	cliPath := strings.TrimSpace(opts.CLIPath)
	if cliPath == "" {
		cliPath = "codex"
		if strings.Contains(kind, "claude") {
			cliPath = "claude"
		} else if strings.Contains(kind, "gemini") {
			cliPath = "gemini"
		}
	}

	// Default models if empty
	if model == "" {
		if strings.Contains(kind, "claude") {
			model = agenttools.DefaultClaudeModel
		} else if strings.Contains(kind, "gemini") {
			model = agenttools.DefaultGeminiModel
		} else {
			model = agenttools.DefaultMetaModel
		}
	}

	return &CLIProvider{
		kind:         kind,
		cliPath:      cliPath,
		model:        model,
		systemPrompt: systemPrompt,
		flags:        append([]string{}, opts.Flags...),
		env:          cloneStringMap(opts.Env),
		toolSpecific: cloneToolSpecific(opts.ToolSpecific),
		logger:       logging.WithComponent(slog.Default(), "meta-cli-"+kind),
	}
}

// SetLogger sets custom logger
func (p *CLIProvider) SetLogger(logger *slog.Logger) {
	p.logger = logging.WithComponent(logger, "meta-cli-"+p.kind)
}

// Name returns provider Kind
func (p *CLIProvider) Name() string {
	return p.kind
}

// TestConnection verifies CLI availability
func (p *CLIProvider) TestConnection(ctx context.Context) error {
	logger := logging.WithTraceID(p.logger, ctx)

	// Check version to verify installation and session (sometimes version check is enough, sometimes need auth status)
	// For now, version check is safe.
	cmd := exec.CommandContext(ctx, p.cliPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("CLI not found or error",
			slog.String("command", p.cliPath),
			slog.String("output", string(output)),
			slog.Any("error", err),
		)
		return fmt.Errorf("%s CLI session check failed: %w (Output: %s)", p.cliPath, err, string(output))
	}

	logger.Info("CLI session verified",
		slog.String("provider", p.kind),
		slog.String("version_output", strings.TrimSpace(string(output))))
	return nil
}

// callExec calls the CLI using agenttools wrapper
func (p *CLIProvider) callExec(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	logger := logging.WithTraceID(p.logger, ctx)
	start := time.Now()

	// Combine system and user prompt for CLI
	fullPrompt := systemPrompt + "\n\n" + userPrompt

	req := agenttools.Request{
		Prompt:          fullPrompt,
		Model:           p.model,
		ReasoningEffort: agenttools.DefaultReasoningEffort,
		Timeout:         DefaultMetaAgentTimeout,
		UseStdin:        true,
		ToolSpecific:    mergeToolSpecific(map[string]interface{}{"docker_mode": false}, p.toolSpecific),
	}

	// Determine provider kind for agenttools
	// If kind contains "claude", use "claude-code" for agenttools mapping
	// If kind contains "codex", use "codex-cli"
	agentToolKind := "codex-cli"
	if strings.Contains(p.kind, "claude") {
		agentToolKind = "claude-code"
	} else if strings.Contains(p.kind, "gemini") {
		agentToolKind = "gemini-cli"
	}

	providerConfig := agenttools.ProviderConfig{
		Kind:     agentToolKind,
		Model:    p.model,
		CLIPath:  p.cliPath,
		Flags:    p.flags,
		ExtraEnv: p.env,
	}

	// Create tool provider
	var toolProvider agenttools.AgentToolProvider
	if agentToolKind == "claude-code" {
		toolProvider = agenttools.NewClaudeProvider(providerConfig)
	} else if agentToolKind == "gemini-cli" {
		toolProvider = agenttools.NewGeminiProvider(providerConfig)
	} else {
		toolProvider = agenttools.NewCodexProvider(providerConfig)
	}

	plan, err := toolProvider.Build(ctx, req)
	if err != nil {
		logger.Error("failed to build exec plan", slog.Any("error", err))
		return "", fmt.Errorf("ExecPlan build failed: %w", err)
	}

	logger.Info("calling CLI",
		slog.String("command", p.cliPath),
		slog.Int("prompt_length", len(fullPrompt)),
		slog.String("model", p.model),
	)

	result := agenttools.Execute(ctx, plan)
	if result.Error != nil {
		logger.Error("CLI call failed",
			slog.String("output", result.Output),
			slog.Int("exit_code", result.ExitCode),
			slog.Any("error", result.Error),
			logging.LogDuration(start),
		)
		return "", fmt.Errorf("CLI call failed: %w (Output: %s)", result.Error, result.Output)
	}

	response := strings.TrimSpace(result.Output)
	logger.Info("CLI call completed",
		slog.Int("response_length", len(response)),
		logging.LogDuration(start),
	)

	return response, nil
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneToolSpecific(in map[string]interface{}) map[string]interface{} {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func mergeToolSpecific(base, override map[string]interface{}) map[string]interface{} {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}
	out := cloneToolSpecific(base)
	if out == nil {
		out = make(map[string]interface{}, len(override))
	}
	for k, v := range override {
		out[k] = v
	}
	return out
}

// Decompose delegates to callExec and extracts response
func (p *CLIProvider) Decompose(ctx context.Context, req *DecomposeRequest) (*DecomposeResponse, error) {
	logger := logging.WithTraceID(p.logger, ctx)

	systemPrompt := decomposeSystemPrompt
	if p.systemPrompt != "" {
		systemPrompt = p.systemPrompt
	}
	userPrompt := buildDecomposeUserPrompt(req)

	resp, err := p.callExec(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	jsonStr := extractJSON(resp)
	yamlStr, err := jsonToYAML(jsonStr)
	if err != nil {
		return nil, fmt.Errorf("failed to convert JSON to YAML: %w\nJSON: %s", err, jsonStr)
	}

	var msg MetaMessage
	if err := yaml.Unmarshal([]byte(yamlStr), &msg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w\nYAML: %s", err, yamlStr)
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
	)

	return &decompose, nil
}

// PlanPatch delegates to callExec
func (p *CLIProvider) PlanPatch(ctx context.Context, req *PlanPatchRequest) (*PlanPatchResponse, error) {
	logger := logging.WithTraceID(p.logger, ctx)

	systemPrompt := planPatchSystemPrompt
	if p.systemPrompt != "" {
		systemPrompt = p.systemPrompt
	}
	userPrompt := buildPlanPatchUserPrompt(req)

	resp, err := p.callExec(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	jsonStr := extractJSON(resp)
	yamlStr, err := jsonToYAML(jsonStr)
	if err != nil {
		return nil, fmt.Errorf("failed to convert JSON to YAML: %w\nJSON: %s", err, jsonStr)
	}

	var msg MetaMessage
	if err := yaml.Unmarshal([]byte(yamlStr), &msg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w\nYAML: %s", err, yamlStr)
	}

	payloadBytes, err := yaml.Marshal(msg.Payload)
	if err != nil {
		return nil, err
	}

	var patch PlanPatchResponse
	if err := yaml.Unmarshal(payloadBytes, &patch); err != nil {
		return nil, fmt.Errorf("failed to parse plan_patch response: %w", err)
	}

	logger.Info("plan_patch completed",
		slog.Int("operations", len(patch.Operations)),
	)

	return &patch, nil
}

// PlanTask delegates to callExec
func (p *CLIProvider) PlanTask(ctx context.Context, prdText string) (*PlanTaskResponse, error) {
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

	resp, err := p.callExec(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

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

// NextAction delegates to callExec
func (p *CLIProvider) NextAction(ctx context.Context, taskSummary *TaskSummary) (*NextActionResponse, error) {
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

	resp, err := p.callExec(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

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

// CompletionAssessment delegates to callExec
func (p *CLIProvider) CompletionAssessment(ctx context.Context, taskSummary *TaskSummary) (*CompletionAssessmentResponse, error) {
	systemPrompt := `You are a Meta-agent evaluating task completion.
Review the Acceptance Criteria and Worker execution results.
Output MUST be a YAML block with type: completion_assessment.
`
	if p.systemPrompt != "" {
		systemPrompt = p.systemPrompt
	}

	acText := ""
	for _, ac := range taskSummary.AcceptanceCriteria {
		acText += fmt.Sprintf("- %s: %s\n", ac.ID, ac.Description)
	}

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

	resp, err := p.callExec(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

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
