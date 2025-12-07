package agenttools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// ExecResult は CLI 実行結果を保持する
type ExecResult struct {
	ExitCode int
	Output   string
	Error    error
}

// GracefulShutdownDelay は SIGTERM 送信後に SIGKILL を送信するまでの待機時間
const GracefulShutdownDelay = 5 * time.Second

// Execute は ExecPlan を実行し、結果を返す
// Docker コンテナ内ではなくホスト上で直接実行する場合に使用
func Execute(ctx context.Context, plan ExecPlan) ExecResult {
	// plan.Timeout が設定されている場合、親コンテキストから独立したコンテキストを作成
	// これにより、親コンテキストのキャンセルに影響されずに実行できる
	execCtx := ctx
	if plan.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(context.Background(), plan.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(execCtx, plan.Command, plan.Args...)

	// Graceful shutdown: コンテキストキャンセル時は SIGTERM → WaitDelay → SIGKILL
	// Go 1.20+ の機能を使用
	cmd.Cancel = func() error {
		return cmd.Process.Signal(syscall.SIGTERM)
	}
	cmd.WaitDelay = GracefulShutdownDelay

	// 環境変数を設定
	if len(plan.Env) > 0 {
		env := cmd.Environ()
		for k, v := range plan.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = env
	}

	// 作業ディレクトリを設定
	if plan.Workdir != "" {
		cmd.Dir = plan.Workdir
	}

	// stdin を設定
	if plan.Stdin != "" {
		cmd.Stdin = strings.NewReader(plan.Stdin)
	}

	output, err := cmd.CombinedOutput()

	result := ExecResult{
		ExitCode: 0,
		Output:   string(output),
		Error:    err,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}

	return result
}
