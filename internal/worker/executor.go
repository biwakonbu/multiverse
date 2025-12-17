package worker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/biwakonbu/agent-runner/internal/agenttools"
	"github.com/biwakonbu/agent-runner/internal/core"
	"github.com/biwakonbu/agent-runner/internal/logging"
	"github.com/biwakonbu/agent-runner/internal/meta"
	"github.com/biwakonbu/agent-runner/pkg/config"
)

type Executor struct {
	Config      config.WorkerConfig
	Sandbox     SandboxProvider
	RepoPath    string
	containerID string // 持続的なコンテナを保持
	logger      *slog.Logger
}

func isClaudeWorkerKind(kind string) bool {
	return kind == "claude-code" || kind == "claude-code-cli"
}

func NewExecutor(cfg config.WorkerConfig, repoPath string) (*Executor, error) {
	sb, err := NewSandboxManager()
	if err != nil {
		return nil, err
	}
	return &Executor{
		Config:      cfg,
		Sandbox:     sb,
		RepoPath:    repoPath,
		containerID: "", // 未初期化
		logger:      logging.WithComponent(slog.Default(), "worker-executor"),
	}, nil
}

// SetLogger sets a custom logger for the executor
func (e *Executor) SetLogger(logger *slog.Logger) {
	e.logger = logging.WithComponent(logger, "worker-executor")
}

// RunWorker executes a worker task
func (e *Executor) RunWorker(ctx context.Context, call meta.WorkerCall, env map[string]string) (*core.WorkerRunResult, error) {
	logger := logging.WithTraceID(e.logger, ctx)

	if e.containerID == "" {
		logger.Error("container not started")
		return nil, fmt.Errorf("container not started: call Start() first")
	}

	// Determine worker type (fallback to config default)
	workerType := call.WorkerType
	if workerType == "" {
		workerType = e.Config.Kind
	}
	if workerType == "" {
		workerType = "codex-cli"
	}

	// Build provider config and request
	reqEnv := mergeEnvMaps(e.Config.Env, call.Env, env)

	// QH-007: Explicitly propagate Codex/Claude session tokens from Host to Container
	// AgentRunner inherits these from Orchestrator (Host), but Docker container needs explicit injection.
	for _, key := range []string{"CODEX_SESSION_TOKEN", "CODEX_API_KEY", "ANTHROPIC_API_KEY"} {
		if val := os.Getenv(key); val != "" {
			if reqEnv == nil {
				reqEnv = make(map[string]string)
			}
			if _, exists := reqEnv[key]; !exists {
				reqEnv[key] = val
			}
		}
	}

	providerCfg := agenttools.ProviderConfig{
		Kind:         workerType,
		CLIPath:      call.CLIPath,
		Model:        call.Model,
		ExtraEnv:     nil, // Env is passed via Request for per-call overrides
		Flags:        call.Flags,
		ToolSpecific: call.ToolSpecific,
	}

	req := agenttools.Request{
		Prompt:          call.Prompt,
		Mode:            call.Mode,
		Model:           call.Model,
		Temperature:     call.Temperature,
		MaxTokens:       call.MaxTokens,
		ReasoningEffort: call.ReasoningEffort,
		Workdir:         call.Workdir,
		Timeout:         0,
		ExtraEnv:        reqEnv,
		Flags:           call.Flags,
		ToolSpecific:    call.ToolSpecific,
		UseStdin:        call.UseStdin,
	}

	// Determine base timeout from config; later overridden by plan.Timeout if set
	timeout := time.Duration(e.Config.MaxRunTimeSec) * time.Second
	if e.Config.MaxRunTimeSec <= 0 {
		timeout = 30 * time.Minute
	}

	plan, err := agenttools.Build(ctx, workerType, providerCfg, req)
	if err != nil {
		return nil, fmt.Errorf("failed to build agent tool plan: %w", err)
	}

	if plan.Timeout > 0 {
		timeout = plan.Timeout
	} else if req.Timeout > 0 {
		timeout = req.Timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var stdin io.Reader
	if plan.Stdin != "" {
		stdin = strings.NewReader(plan.Stdin)
	}

	cmd := []string{plan.Command}
	cmd = append(cmd, plan.Args...)

	if len(plan.Env) > 0 {
		envPrefix := buildEnvPrefix(plan.Env)
		cmd = append(envPrefix, cmd...)
	}

	containerID := e.containerID
	containerLabel := containerID
	if len(containerID) > 12 {
		containerLabel = containerID[:12]
	}

	logger.Info("executing worker command",
		slog.String("container_id", containerLabel),
		slog.Int("prompt_length", len(call.Prompt)),
		slog.Float64("timeout_sec", timeout.Seconds()),
	)
	logger.Debug("worker command details",
		slog.String("prompt", call.Prompt),
		slog.Any("cmd", cmd),
	)

	start := time.Now()
	exitCode, output, execErr := e.Sandbox.Exec(ctx, containerID, cmd, stdin)
	finish := time.Now()

	res := &core.WorkerRunResult{
		ID:         fmt.Sprintf("run-%d", start.Unix()),
		StartedAt:  start,
		FinishedAt: finish,
		ExitCode:   exitCode,
		RawOutput:  output,
		Summary:    "Worker executed",
		Error:      execErr,
	}

	// Capture artifacts if execution was successful (or even if failed, we might want to see changes)
	// QH-008: Track modified files
	if artifacts, err := e.captureArtifacts(ctx, containerID); err == nil && len(artifacts) > 0 {
		res.Artifacts = artifacts
		logger.Info("artifacts detected", slog.Int("count", len(artifacts)))
	}

	durationMs := float64(finish.Sub(start).Milliseconds())
	if execErr != nil {
		logger.Error("worker execution failed",
			slog.Int("exit_code", exitCode),
			slog.Float64("duration_ms", durationMs),
			slog.Any("error", execErr),
		)
	} else {
		logger.Info("worker execution completed",
			slog.Int("exit_code", exitCode),
			slog.Int("output_length", len(output)),
			slog.Float64("duration_ms", durationMs),
		)
		logger.Debug("worker output", slog.String("output", output))
	}

	return res, nil
}

// Start starts a persistent container for the task
func (e *Executor) Start(ctx context.Context) error {
	logger := logging.WithTraceID(e.logger, ctx)

	if e.containerID != "" {
		logger.Warn("container already started", slog.String("container_id", e.containerID[:12]))
		return fmt.Errorf("container already started (ID: %s)", e.containerID)
	}

	if isClaudeWorkerKind(e.Config.Kind) {
		if err := e.verifyClaudeSession(ctx); err != nil {
			logger.Error("claude session verification failed",
				slog.Any("error", err),
				slog.String("hint", "claude login で認証してください"),
			)
			return fmt.Errorf("claude code session missing: %w", err)
		}
	} else if e.Config.Kind == "codex-cli" || e.Config.Kind == "" {
		if err := e.verifyCodexSession(ctx); err != nil {
			logger.Error("codex CLI session verification failed",
				slog.Any("error", err),
				slog.String("hint", "codex login で認証するか ~/.codex/auth.json を用意してください"),
			)
			return fmt.Errorf("Codex CLI session missing: %w", err)
		}
	}

	image := e.Config.DockerImage
	if image == "" {
		if isClaudeWorkerKind(e.Config.Kind) {
			image = "ghcr.io/biwakonbu/agent-runner-claude:latest"
		} else {
			image = "ghcr.io/biwakonbu/agent-runner-codex:latest"
		}
	}

	// Resolve RepoPath to absolute path
	repoPath := e.RepoPath
	if repoPath == "" {
		repoPath = "."
	}
	absRepo, err := filepath.Abs(repoPath)
	if err != nil {
		logger.Error("failed to get absolute path", slog.String("repo_path", repoPath), slog.Any("error", err))
		return fmt.Errorf("failed to get absolute path for %s: %w", repoPath, err)
	}
	repoPath = absRepo

	logger.Info("starting container",
		slog.String("image", image),
		slog.String("repo_path", repoPath),
	)

	start := time.Now()
	// Pass internal auth path via env if configured
	startEnv := make(map[string]string)
	if isClaudeWorkerKind(e.Config.Kind) && e.Config.AuthPath != "" {
		startEnv["__INTERNAL_CLAUDE_AUTH_PATH"] = e.Config.AuthPath
	}

	containerID, err := e.Sandbox.StartContainer(ctx, image, repoPath, startEnv)
	if err != nil {
		logger.Error("failed to start container",
			slog.String("image", image),
			slog.Any("error", err),
			logging.LogDuration(start),
		)
		return fmt.Errorf("failed to start container: %w", err)
	}

	e.containerID = containerID

	containerLabel := containerID
	if len(containerID) > 12 {
		containerLabel = containerID[:12]
	}
	logger.Info("container started",
		slog.String("container_id", containerLabel),
		logging.LogDuration(start),
	)
	return nil
}

// verifyCodexSession は Codex CLI セッションの存在を検証する
func (e *Executor) verifyCodexSession(ctx context.Context) error {
	// 1. ~/.codex/auth.json の存在確認
	homeDir, err := os.UserHomeDir()
	if err == nil {
		codexAuthPath := filepath.Join(homeDir, ".codex", "auth.json")
		if _, err := os.Stat(codexAuthPath); err == nil {
			// auth.json が存在する
			return nil
		}
	}

	// 2. CODEX_API_KEY or CODEX_SESSION_TOKEN environment variable check
	if os.Getenv("CODEX_API_KEY") != "" || os.Getenv("CODEX_SESSION_TOKEN") != "" {
		return nil
	}

	// 3. codex --version で CLI が利用可能か確認
	cmd := exec.CommandContext(ctx, "codex", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("codex CLI not found or not authenticated: %w (出力: %s)", err, string(output))
	}

	// CLI は存在するが、セッション情報が不明
	return fmt.Errorf("codex CLI セッションが検出できません。`codex login` で認証を完了してください。出力: %s", strings.TrimSpace(string(output)))
}

// verifyClaudeSession checks for Claude Code session existence
func (e *Executor) verifyClaudeSession(ctx context.Context) error {
	// 1. Check for Config path
	homeDir, err := os.UserHomeDir()
	foundConfig := false
	if err == nil {
		configPath := ""
		if e.Config.AuthPath != "" {
			// user configured path
			if filepath.IsAbs(e.Config.AuthPath) {
				configPath = e.Config.AuthPath
			} else {
				configPath = filepath.Join(homeDir, e.Config.AuthPath)
			}
		} else {
			// default path: ~/.config/claude
			configPath = filepath.Join(homeDir, ".config", "claude")
		}

		if _, err := os.Stat(configPath); err == nil {
			foundConfig = true
		}
	}

	if foundConfig {
		return nil
	}

	// 2. Check claude --version
	cmd := exec.CommandContext(ctx, "claude", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("claude CLI not found: %w", err)
	}

	// CLI exists but no auth file found.
	// Since we rely on mounting the auth directory, we strictly need the directory.
	return fmt.Errorf("claude authentication directory not found. Please run `claude login`. (claude version: %s)", strings.TrimSpace(string(output)))
}

// Stop stops the persistent container
func (e *Executor) Stop(ctx context.Context) error {
	logger := logging.WithTraceID(e.logger, ctx)

	if e.containerID == "" {
		logger.Warn("no container to stop")
		return fmt.Errorf("no container to stop")
	}

	// Store containerID before clearing (for error message)
	containerID := e.containerID

	containerLabel := containerID
	if len(containerID) > 12 {
		containerLabel = containerID[:12]
	}
	logger.Info("stopping container", slog.String("container_id", containerLabel))

	// Clear containerID first to prevent resource leak
	// even if StopContainer fails
	e.containerID = ""

	start := time.Now()
	err := e.Sandbox.StopContainer(ctx, containerID)
	if err != nil {
		logger.Error("failed to stop container",
			slog.String("container_id", containerLabel),
			slog.Any("error", err),
			logging.LogDuration(start),
		)
		return fmt.Errorf("failed to stop container: %w", err)
	}

	logger.Info("container stopped",
		slog.String("container_id", containerLabel),
		logging.LogDuration(start),
	)
	return nil
}

// mergeEnvMaps merges multiple env maps left-to-right, ignoring nil entries.
func mergeEnvMaps(envs ...map[string]string) map[string]string {
	result := map[string]string{}
	for _, env := range envs {
		for k, v := range env {
			result[k] = v
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// buildEnvPrefix renders env vars as ["env", "KEY=VAL", ...] for container exec.
func buildEnvPrefix(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	prefix := []string{"env"}
	for k, v := range env {
		val := v
		if strings.HasPrefix(v, "env:") {
			val = os.Getenv(strings.TrimPrefix(v, "env:"))
		}
		prefix = append(prefix, fmt.Sprintf("%s=%s", k, val))
	}
	return prefix
}

// captureArtifacts detects modified files using git status
func (e *Executor) captureArtifacts(ctx context.Context, containerID string) ([]string, error) {
	// Simple git status check
	// porcelain format provides stable output: "M  file", "?? file", etc.
	cmd := []string{"git", "status", "--porcelain"}

	// Create short timeout context for artifact detection
	dtCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, output, err := e.Sandbox.Exec(dtCtx, containerID, cmd, nil)
	if err != nil {
		// If fails (not a git repo, or git not installed), we just log and return empty
		// e.logger.Debug("failed to capture artifacts (git status)", slog.Any("error", err))
		return nil, nil // Non-critical failure
	}

	var artifacts []string
	validCodes := map[byte]bool{
		'M': true, // Modified
		'A': true, // Added
		'D': true, // Deleted
		'R': true, // Renamed
		'C': true, // Copied
		'?': true, // Untracked
	}

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) < 4 {
			continue
		}

		// Porcelain format: XY PATH
		// X = staging status, Y = worktree status
		x := trimmed[0]
		y := trimmed[1]

		if validCodes[x] || validCodes[y] {
			// Extract path (handle potentially quoted paths if needed, but simple split for now)
			// Status is usually fitst 2 chars, then space, then path
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				// Rejoin parts in case of spaces in filename (though porcelain quotes them usually)
				// For simple MVP: just take the rest
				path := strings.TrimSpace(trimmed[2:])
				artifacts = append(artifacts, path)
			}
		}
	}

	return artifacts, nil
}
