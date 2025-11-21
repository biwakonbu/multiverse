# Worker Package - Worker 実行とサンドボックス管理

このパッケージはWorker CLI（例：Codex CLI）を安全に実行するためのDocker サンドボックス環境を管理します。

## 概要

- **executor.go**: Worker CLI実行の抽象化、Runner からの呼び出しインターフェース
- **sandbox.go**: Docker API統合、コンテナ生成・実行・クリーンアップ、マウント管理

## アーキテクチャ

### 関係図

```
Runner (FSM)
  ↓ RunWorker(prompt, env)
Executor
  ↓ StartContainer(), Exec(), StopContainer()
SandboxManager
  ↓ Docker API (docker/client)
Docker Daemon
  ↓ (コンテナ起動)
agentrunner-codex:latest イメージ
```

## Executor インターフェース

**定義** (executor.go:13-29):

```go
type Executor struct {
    Config   config.WorkerConfig  // Docker image 等の設定
    Sandbox  *SandboxManager      // Docker操作
    RepoPath string               // リポジトリの絶対パス
}

// Runner が呼び出すメソッド
func (e *Executor) RunWorker(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error)
```

### RunWorker() フロー

```
1. StartContainer()
   ├─ リポジトリをバインドマウント（/workspace/project）
   ├─ 認証情報をマウント（~/.codex/auth.json）
   └─ 環境変数注入
2. Exec()
   ├─ Codex CLI コマンド実行
   ├─ 標準出力・標準エラー収集
   └─ 終了コード取得
3. StopContainer()
   └─ コンテナ強制停止（タイムアウト0秒）
```

**実装** (executor.go:31-111):

```go
func (e *Executor) RunWorker(ctx context.Context, prompt string, env map[string]string) (*core.WorkerRunResult, error) {
    // 1. コンテナ起動
    containerID, err := e.Sandbox.StartContainer(ctx, image, repoPath, env)
    defer e.Sandbox.StopContainer(ctx, containerID) // 必ず停止

    // 2. Codex CLI コマンド構築
    cmd := []string{
        "codex", "exec",
        "--sandbox", "workspace-write",
        "--json",
        "--cwd", "/workspace/project",
        prompt,  // Meta-agentが生成したプロンプト
    }

    // 3. 実行
    exitCode, output, err := e.Sandbox.Exec(ctx, containerID, cmd)

    // 4. 結果をWorkerRunResultにラップして返す
    res := &core.WorkerRunResult{
        ID:         fmt.Sprintf("run-%d", start.Unix()),
        StartedAt:  start,
        FinishedAt: finish,
        ExitCode:   exitCode,
        RawOutput:  output,
        Summary:    "Worker executed",
        Error:      err,
    }
    return res, nil
}
```

## SandboxManager（Docker統合）

**定義** (sandbox.go:17-27):

```go
type SandboxManager struct {
    cli *client.Client  // Docker API クライアント
}

func NewSandboxManager() (*SandboxManager, error) {
    cli, err := client.NewClientWithOpts(
        client.FromEnv,                          // 環境変数から認証情報取得
        client.WithAPIVersionNegotiation(),      // サーバーバージョンに自動調整
    )
}
```

### 1. StartContainer - コンテナ起動

**処理フロー** (sandbox.go:29-87):

```go
func (s *SandboxManager) StartContainer(ctx context.Context, image string, repoPath string, env map[string]string) (string, error) {
```

#### 環境変数の準備

```go
var envSlice []string
for k, v := range env {
    envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
}
```

- Runner から渡された `env` マップを `KEY=VALUE` スライスに変換

#### バインドマウント設定

```go
mounts := []mount.Mount{
    {
        Type:   mount.TypeBind,
        Source: repoPath,              // ホスト側：リポジトリ絶対パス
        Target: "/workspace/project",  // コンテナ側：固定パス
    },
}
```

**重要**: Source は**絶対パス**が必須（相対パスはDocker APIエラー）

#### 認証情報マウント（オプション）

```go
homeDir, err := os.UserHomeDir()
codexAuthPath := filepath.Join(homeDir, ".codex", "auth.json")
if _, err := os.Stat(codexAuthPath); err == nil {
    // ~/.codex/auth.json が存在する場合、read-only でマウント
    mounts = append(mounts, mount.Mount{
        Type:     mount.TypeBind,
        Source:   codexAuthPath,
        Target:   "/root/.codex/auth.json",
        ReadOnly: true,  // コンテナからの書き込み禁止
    })
}
```

**フォールバック**:

```go
if codexAPIKey := os.Getenv("CODEX_API_KEY"); codexAPIKey != "" {
    envSlice = append(envSlice, fmt.Sprintf("CODEX_API_KEY=%s", codexAPIKey))
}
```

- auth.json が存在しない場合、環境変数 `CODEX_API_KEY` を コンテナに注入

#### コンテナ生成・起動

```go
resp, err := s.cli.ContainerCreate(ctx, &container.Config{
    Image:      image,                          // "agent-runner-codex:latest" 等
    Tty:        true,
    Env:        envSlice,
    Cmd:        []string{"tail", "-f", "/dev/null"}, // 起動時コマンド（待機）
    WorkingDir: "/workspace/project",
}, &container.HostConfig{
    Mounts: mounts,
}, nil, nil, "")
```

- **Cmd**: `tail -f /dev/null` で無限待機（Exec で別コマンドを実行するため）

起動:

```go
s.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
return resp.ID, nil
```

### 2. Exec - コマンド実行

**処理フロー** (sandbox.go:89-129):

```go
func (s *SandboxManager) Exec(ctx context.Context, containerID string, cmd []string) (int, string, error)
```

#### ExecConfig 設定

```go
execConfig := types.ExecConfig{
    Cmd:          cmd,                // ["codex", "exec", "--sandbox", ...]
    AttachStdout: true,               // 標準出力キャプチャ
    AttachStderr: true,               // 標準エラーキャプチャ
    Tty:          false,              // Ttyなし（stdout/stderr分離用）
    WorkingDir:   "/workspace/project",
}
```

#### ExecCreate & ExecAttach

```go
resp, err := s.cli.ContainerExecCreate(ctx, containerID, execConfig)
hijacked, err := s.cli.ContainerExecAttach(ctx, resp.ID, types.ExecStartCheck{})
defer hijacked.Close()
```

- `ExecCreate`: コマンドを登録
- `ExecAttach`: コマンド実行・出力ストリーム接続

#### 出力収集

```go
var outBuf, errBuf bytes.Buffer
_, err = stdcopy.StdCopy(&outBuf, &errBuf, hijacked.Reader)
// stdcopy: Docker の特殊フォーマット（stdout/stderr タグ付き）をパース
```

**stdcopy について**:
- Docker の `Exec` は stdout/stderr を特殊フォーマット（frame format）で返す
- `stdcopy.StdCopy()` がこれを処理して、outBuf/errBuf に分離

#### 終了コード取得

```go
inspectResp, err := s.cli.ContainerExecInspect(ctx, resp.ID)
return inspectResp.ExitCode, output, nil
```

### 3. StopContainer - コンテナ停止

**実装** (sandbox.go:131-134):

```go
func (s *SandboxManager) StopContainer(ctx context.Context, containerID string) error {
    timeout := 0 // Force kill（SIGKILL）
    return s.cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
}
```

- `Timeout: 0` で即座に kill（graceful shutdown なし）
- **理由**: タスク実行時の安全性のため、コンテナが確実に停止することを優先

## マウント戦略

### リポジトリの永続性

**バインドマウント使用**:
```
ホスト: /Users/biwakonbu/github/agent-runner
コンテナ: /workspace/project
```

- **利点**: コンテナ内で編集したファイルがホスト側に永続化
- Worker実行後、コンテナを停止しても ファイル変更は保持される

**タスク実行フロー**:
```
1. Worker実行 → コンテナ内でコード編集
2. コンテナ停止
3. 次のWorker実行 → 新しいコンテナで前回の編集が見える
4. ループ継続
```

### 認証情報のセキュリティ

**~/codex/auth.json**:
- `ReadOnly: true` でマウント → コンテナからの修正不可
- コンテナ内で read-only ファイルシステムに配置

**CODEX_API_KEY 環境変数**:
- auth.json が存在しない場合のフォールバック
- ホスト側の環境変数を直接注入

## Docker Image: agentrunner-codex

**Dockerfile** (sandbox/Dockerfile):
```dockerfile
FROM python:3.11-slim
RUN apt-get update && apt-get install -y nodejs git
RUN pip install codex-cli
WORKDIR /workspace/project
CMD ["tail", "-f", "/dev/null"]
```

**要件**:
- Python 3.11+
- Node.js（プロジェクト要件による）
- Git
- Codex CLI インストール済み

**ビルド**:
```bash
docker build -t agent-runner-codex:latest sandbox/
```

## 既知の問題と回避策

### 問題1: 相対パスでのマウントエラー

**症状**:
```
Error: invalid mount path: '.' mount path must be absolute
```

**原因**:
Docker API は Source に絶対パスを要求。runner.go で `repoPath = "."` をそのまま渡すとエラー。

**回避策** (executor.go:73-77):
```go
repoPath := e.RepoPath
if repoPath == "" {
    absRepo, _ := filepath.Abs(".")
    repoPath = absRepo
}
```

- Executor初期化時に `repoPath` を絶対パスに解決
- Runner から渡された RepoPath が既に絶対パス（runner.go:61-62で解決済み）

**推奨**: Runner で常に絶対パスを使用することを強制

### 問題2: Docker Daemon への接続エラー

**症状**:
```
Error: cannot connect to the Docker daemon
```

**原因**:
- Docker デーモンが起動していない
- ソケットの権限エラー（Unix）

**対応**:
```bash
# Docker 起動確認
docker ps

# 権限問題（Linux）
sudo usermod -aG docker $USER
```

### 問題3: イメージが存在しない

**症状**:
```
Error: image not found: agent-runner-codex:latest
```

**対応**:
```bash
docker build -t agent-runner-codex:latest sandbox/
```

## パフォーマンス最適化

### 1. イメージプル（MVP では未実装）

**現在**: イメージがローカルに存在することを前提

**今後改善案**:
```go
// ImagePull（実装されていない）
_, err := s.cli.ImagePull(ctx, image, types.ImagePullOptions{})
```

- リモートレジストリから自動取得

### 2. コンテナライフサイクル

**現在**: RunWorker ごとに start/stop
- **効率**: 低（毎回起動オーバーヘッド）
- **シンプリシティ**: 高（ライフサイクル管理なし）
- **安全性**: 高（リソースリークなし）

**将来最適化**:
- Task ごとに1コンテナ（persistent）
- RunWorker では exec のみ
- Task 完了時に stop

### 3. 環境変数注入の最適化

**現在**: 毎回 `env` マップを `KEY=VALUE` スライスに変換

**最適化案**:
```go
// Executor 初期化時に一度だけ変換
type Executor struct {
    ...
    envSlice []string
}
```

## トラブルシューティング

### Step 1: ローカルでテスト
```bash
docker run -it --rm agent-runner-codex:latest bash
# コンテナ内でコマンドテスト
codex exec --help
```

### Step 2: Docker ログ確認
```bash
docker logs <container-id>
```

### Step 3: Executor をモック置き換え
```go
// internal/mock/worker.go
func NewMockExecutor() *mock.WorkerExecutor {
    // テスト用モック実装
}
```

統合テストで mock を使用して Docker 依存を除去。

## コード例：タスク実行フロー

```go
// Runner から呼び出される
res, err := e.RunWorker(ctx, "ユーザー登録機能を実装", map[string]string{})

// Executor 内部処理
{
    // 1. コンテナ起動（リポジトリマウント）
    containerID, _ := e.Sandbox.StartContainer(ctx, "agent-runner-codex:latest", "/Users/biwakonbu/github/agent-runner", nil)
    defer e.Sandbox.StopContainer(ctx, containerID)

    // 2. Codex CLI 実行
    cmd := []string{
        "codex", "exec",
        "--sandbox", "workspace-write",
        "--json",
        "--cwd", "/workspace/project",
        "ユーザー登録機能を実装",
    }
    exitCode, output, err := e.Sandbox.Exec(ctx, containerID, cmd)

    // 3. 結果をラップ
    res := &core.WorkerRunResult{
        ExitCode:   exitCode,
        RawOutput:  output,
        Error:      err,
    }
}

// Runner が res を タスク履歴に記録
taskCtx.WorkerRuns = append(taskCtx.WorkerRuns, *res)
```

## 関連ドキュメント

- [core/CLAUDE.md](../core/CLAUDE.md): Runner FSM と TaskContext
- [meta/CLAUDE.md](../meta/CLAUDE.md): Meta-agent通信の詳細
- `/docs/AgentRunner-architecture.md`: アーキテクチャ全体
- `sandbox/Dockerfile`: Docker イメージ定義
- `TESTING.md`: テストベストプラクティス
- `CODEX_TEST_README.md`: Codex 統合ガイド
