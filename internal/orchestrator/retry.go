package orchestrator

import (
	"math"
	"time"
)

// RetryPolicy はタスク失敗時のリトライポリシーを定義する
type RetryPolicy struct {
	MaxAttempts   int           // 最大試行回数（デフォルト: 3）
	BackoffBase   time.Duration // バックオフ基準時間（デフォルト: 5秒）
	BackoffMax    time.Duration // バックオフ最大時間（デフォルト: 5分）
	BackoffFactor float64       // バックオフ乗数（デフォルト: 2.0）
	RequireHuman  bool          // 最大試行後に人間判断を要求するか
}

// DefaultRetryPolicy はデフォルトのリトライポリシーを返す
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:   3,
		BackoffBase:   5 * time.Second,
		BackoffMax:    5 * time.Minute,
		BackoffFactor: 2.0,
		RequireHuman:  true,
	}
}

// CalculateBackoff は次のリトライまでの待機時間を計算する
// attemptNumber は 1 から始まる試行回数
func (p *RetryPolicy) CalculateBackoff(attemptNumber int) time.Duration {
	if attemptNumber <= 0 {
		attemptNumber = 1
	}

	// 指数バックオフ: base * factor^(attempt-1)
	backoff := float64(p.BackoffBase) * math.Pow(p.BackoffFactor, float64(attemptNumber-1))

	// 最大値でキャップ
	if backoff > float64(p.BackoffMax) {
		backoff = float64(p.BackoffMax)
	}

	return time.Duration(backoff)
}

// ShouldRetry はリトライすべきかを判定する
// attemptNumber は現在の試行回数（1から始まる）
func (p *RetryPolicy) ShouldRetry(attemptNumber int) bool {
	return attemptNumber < p.MaxAttempts
}

// NextAction は次のアクションを決定する
type NextAction string

const (
	NextActionRetry   NextAction = "RETRY"   // リトライする
	NextActionBacklog NextAction = "BACKLOG" // バックログに追加（人間判断）
	NextActionFail    NextAction = "FAIL"    // 失敗としてマーク
)

// DetermineNextAction は失敗後の次のアクションを決定する
func (p *RetryPolicy) DetermineNextAction(attemptNumber int) NextAction {
	if p.ShouldRetry(attemptNumber) {
		return NextActionRetry
	}

	if p.RequireHuman {
		return NextActionBacklog
	}

	return NextActionFail
}
