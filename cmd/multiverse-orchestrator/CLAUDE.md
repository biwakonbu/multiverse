# multiverse-orchestrator - Orchestrator CLI（実装予定）

このディレクトリは multiverse Orchestrator の CLI エントリポイントを提供します（現在未実装）。

## 責務（予定）

- **Worker 管理**: Worker プロセスの起動・監視・終了
- **IPC Queue 処理**: Queue からジョブを取得し Worker に割り当て
- **AgentRunner Core 連携**: タスクごとに AgentRunner Core を起動

## 予定アーキテクチャ

```
┌─────────────────────────────────────┐
│  multiverse-ide                    │
│  (Task 作成、スケジュール)          │
└──────────────┬──────────────────────┘
               │ IPC Queue (filesystem)
┌──────────────▼──────────────────────┐
│  multiverse-orchestrator           │
│  - Queue 監視                       │
│  - Worker Pool 管理                 │
│  - AgentRunner Core 起動            │
└──────────────┬──────────────────────┘
               │ subprocess / Docker
┌──────────────▼──────────────────────┐
│  AgentRunner Core (agent-runner)   │
│  + Worker (Docker Sandbox)         │
└─────────────────────────────────────┘
```

## 予定機能

### Queue 監視

```go
// 擬似コード
for {
    jobs := queue.ListJobs(poolID)
    for _, job := range jobs {
        go processJob(job)
    }
    time.Sleep(pollInterval)
}
```

### Worker Pool 管理

- Worker Pool ごとに並列度を制限
- アイドル Worker の再利用
- タイムアウト処理

### AgentRunner Core 起動

```go
// 擬似コード
cmd := exec.Command("agent-runner")
cmd.Stdin = taskYAML
cmd.Run()
```

## 現状

- **ディレクトリのみ存在**: 実装は未着手
- **代替手段**: IDE から直接 Scheduler を呼び出し、IPC Queue にジョブを投入

## 実装予定

### Phase 1: 基本実装

- [ ] main.go: CLI エントリポイント
- [ ] Queue 監視ループ
- [ ] AgentRunner Core サブプロセス起動

### Phase 2: 高度な機能

- [ ] Worker Pool 並列度制御
- [ ] リトライロジック
- [ ] 結果の IPC Queue 書き込み

### Phase 3: 運用機能

- [ ] ヘルスチェック
- [ ] メトリクス出力
- [ ] graceful shutdown

## 関連ドキュメント

- [../../internal/orchestrator/CLAUDE.md](../../internal/orchestrator/CLAUDE.md): TaskStore, Scheduler
- [../agent-runner/CLAUDE.md](../agent-runner/CLAUDE.md): AgentRunner Core CLI
- [../../PRD.md](../../PRD.md): multiverse IDE v0.1 要件
