# CLAUDE.md

このファイルはこのリポジトリでClaudeコードを操作する際のガイダンスを提供します。

## プロジェクト概要

**agent-runner** は、開発タスクを自動実行するメタエージェント・オーケストレーションレイヤーです。以下を組み合わせています：
- **Meta-agent**: LLM（GPT-4 Turbo）がタスクを計画・評価
- **AgentRunner Core**: タスク状態と実行フローを管理するエンジン
- **Worker Agents**: 実際の開発作業（コーディング、テスト実行）を行うCLIツール（例：Codex CLI）

Docker サンドボックス環境で安全・再現性高く実行され、全ての実行履歴はMarkdown形式のTask Note（`.agent-runner/task-*.md`）として保存されます。

## よく使うコマンド

### ビルド
```bash
# バイナリをビルド
go build ./cmd/agent-runner

# 出力ファイル名を指定
go build -o agent-runner ./cmd/agent-runner
```

### テスト
```bash
# 全テストを実行
go test ./...

# 特定パッケージのテストを実行
go test ./internal/core -run TestRunner_Properties

# 統合テストを実行
go test ./test/integration/...

# タイムアウトを指定
go test -timeout 30s ./...
```

### 実行
```bash
# 基本的な実行（標準入力からYAMLを読み込む）
./agent-runner < task.yaml

# 設定ファイルをパイプ
cat task.yaml | go run cmd/agent-runner/main.go
```

### Docker（Worker実行環境）
```bash
# Codex worker のランタイムをビルド
docker build -t agent-runner-codex:latest sandbox/

# Codex統合テストを実行
./run_codex_test.sh
```

## アーキテクチャ

### 三層構造
```
Meta-agent (LLM)
    ↕ YAML プロトコル ↕
AgentRunner Core (状態 + TaskContext)
    ↕ Exec + Docker ↕
Worker Agents (Codex CLI等)
```

### タスク状態機械（FSM）
```
PENDING → PLANNING → RUNNING → VALIDATING → COMPLETE
                                              ↓
                                            FAILED
```

**主要な遷移**:
- `PENDING → PLANNING`: Meta-agentがPRDから受け入れ条件（AC）を生成
- `PLANNING → RUNNING`: Meta-agentがworker実行を要求
- `RUNNING → VALIDATING`: Worker実行が完了、Meta-agentが完了を評価
- `VALIDATING → COMPLETE/FAILED`: タスク完了判定を決定

### 主要パッケージ

#### `/internal/core`
- **runner.go**: タスクFSMのオーケストレーション、TaskContextの管理
- **context.go**: TaskContext（PRD、AC、worker実行、meta呼び出し）とTaskStateを定義
- `gopter`を使用した状態不変条件のプロパティベーステスト

#### `/internal/meta`
- **client.go**: Meta-agent推論用のOpenAI API通信
- **protocol.go**: YAML メッセージ構造（`PlanTaskResponse`、`NextActionResponse`等）
- モック実装に対応（`kind: "mock"`）

#### `/internal/worker`
- **executor.go**: Worker CLI実行の抽象化
- **sandbox.go**: Docker API管理（コンテナ作成、exec、クリーンアップ）
  - リポジトリは`/workspace/project`にマウント
  - 認証情報を自動マウント（`~/.codex/auth.json`）

#### `/internal/note`
- **writer.go**: TaskContextからMarkdown Task Noteを生成
- Go `text/template`によるテンプレートベース出力

#### `/pkg/config`
- **config.go**: タスク設定構造体（YAMLパース）

## タスク YAML 形式

```yaml
version: 1

task:
  id: "TASK-123"              # オプション（省略時は自動生成）
  title: "タスク名"           # オプション
  repo: "."                   # 作業対象リポジトリ（相対パスまたは絶対パス）

  prd:
    path: "./docs/prd.md"     # PRDファイルパス、または
    text: |                   # PRDテキストを直接埋め込み
      要件定義...

  test:
    command: "npm test"       # オプション（自動テストコマンド）
    cwd: "./"                 # オプション（テスト実行ディレクトリ）

runner:
  meta:
    kind: "openai-chat"       # または "mock"
    model: "gpt-4-turbo"      # LLMモデル名

  worker:
    kind: "codex-cli"
    docker_image: "agent-runner-codex:latest"
    max_run_time_sec: 1800    # 実行タイムアウト
    env:
      CODEX_API_KEY: "env:CODEX_API_KEY"  # 環境変数を参照
```

**参考**: [sample_task_go.yaml](sample_task_go.yaml)、[test_codex_task.yaml](test_codex_task.yaml)

## 重要な設計パターン

### 依存性注入（Dependency Injection）
`Runner`構造体はインターフェース（`MetaClient`、`WorkerExecutor`、`NoteWriter`）を受け入れるため、統合依存性なしでモックベースのテストが可能です。

```go
type Runner struct {
    Config *config.TaskConfig
    Meta   MetaClient        // インターフェース
    Worker WorkerExecutor    // インターフェース
    Note   NoteWriter        // インターフェース
}
```

### TaskContext の伝播
実行状態（PRD、受け入れ条件、worker結果、meta呼び出し、タイミング）全てが単一の`TaskContext`構造体で保持され、FSMを通じて伝播します。これにより完全な再現性と監査証跡（Task Noteに記録）が実現されます。

### Sandbox 隔離
- タスク1つあたり1つのDockerコンテナ（`/workspace/project`マウントポイント）
- 完了またはエラー時の自動クリーンアップ
- 環境変数注入と認証情報マウントをサポート

## 既知の課題

### 相対パスの解決
タスク設定で相対パス`"."`を使用するとDockerマウントエラーが発生します：
```
invalid mount path: '.' mount path must be absolute
```

**対応**: 絶対パスを使用するか、worker実行前にリポジトリパスを絶対パスに解決してください（`worker/executor.go`で`filepath.Abs`を使用）。

## 環境変数

```bash
# Meta-agent用（OpenAI API）
export OPENAI_API_KEY="sk-..."

# Worker agent用（auth.json未使用の場合）
export CODEX_API_KEY="..."
```

## 関連ドキュメント

- **[GEMINI.md](GEMINI.md)**: プロジェクト概要と背景
- **[TESTING.md](TESTING.md)**: テストベストプラクティス
- **[CODEX_TEST_README.md](CODEX_TEST_README.md)**: Codex統合ガイド
- **[docs/AgentRunner-architecture.md](docs/AgentRunner-architecture.md)**: アーキテクチャ詳細仕様
- **[docs/agentrunner-spec-v1.md](docs/agentrunner-spec-v1.md)**: MVP/v1仕様書
- **[docs/AgentRunner-impl-design-v1.md](docs/AgentRunner-impl-design-v1.md)**: Go実装設計

## 開発ノート

- **言語**: コメントは日本語、関数・変数名は英語
- **テスト**: 依存性注入とモックを使用；プロパティベーステストで不変条件を検証
- **ロギング**: 現在`fmt.Printf`を使用（今後`slog`への移行を検討）
- **依存関係**: Go 1.24.0以上、Docker、OpenAI API
