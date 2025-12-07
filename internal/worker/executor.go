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

	// Verify Codex CLI session before starting container
	if err := e.verifyCodexSession(ctx); err != nil {
		logger.Error("codex CLI session verification failed",
			slog.Any("error", err),
			slog.String("hint", "codex login で認証するか ~/.codex/auth.json を用意してください"),
		)
		// セッションなしは致命的とみなし、IDE 側に伝播させる
		return fmt.Errorf("Codex CLI セッションがありません: %w", err)
	}

	image := e.Config.DockerImage
	if image == "" {
		image = "ghcr.io/biwakonbu/agent-runner-codex:latest" // Default
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
	containerID, err := e.Sandbox.StartContainer(ctx, image, repoPath, nil)
	if err != nil {
		logger.Error("failed to start container",
			slog.String("image", image),
			slog.Any("error", err),
			logging.LogDuration(start),
		)
		return fmt.Errorf("failed to start container: %w", err)
	}

	e.containerID = containerID
	logger.Info("container started",
		slog.String("container_id", containerID[:12]),
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

	// 2. CODEX_API_KEY 環境変数の確認
	if os.Getenv("CODEX_API_KEY") != "" {
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

// Stop stops the persistent container
func (e *Executor) Stop(ctx context.Context) error {
	logger := logging.WithTraceID(e.logger, ctx)

	if e.containerID == "" {
		logger.Warn("no container to stop")
		return fmt.Errorf("no container to stop")
	}

	// Store containerID before clearing (for error message)
	containerID := e.containerID
	logger.Info("stopping container", slog.String("container_id", containerID[:12]))

	// Clear containerID first to prevent resource leak
	// even if StopContainer fails
	e.containerID = ""

	start := time.Now()
	err := e.Sandbox.StopContainer(ctx, containerID)
	if err != nil {
		logger.Error("failed to stop container",
			slog.String("container_id", containerID[:12]),
			slog.Any("error", err),
			logging.LogDuration(start),
		)
		return fmt.Errorf("failed to stop container: %w", err)
	}

	logger.Info("container stopped",
		slog.String("container_id", containerID[:12]),
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
