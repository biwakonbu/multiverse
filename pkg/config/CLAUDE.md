# Config Package - YAML 設定スキーマ

このパッケージはタスク実行の全設定を定義し、YAMLファイルから Go構造体へのパース・検証を担当します。設定スキーマはプロジェクト全体のインターフェース仕様として機能します。

## 概要

- **config.go**: YAML 構造体定義（TaskConfig と その階層構造）
- **config_test.go**: パース・検証・エラーハンドリングテスト（320行）

## TaskConfig 階層構造

### 全体像

```go
type TaskConfig struct {
    Version int          `yaml:"version"`
    Task    TaskDetails  `yaml:"task"`
    Runner  RunnerConfig `yaml:"runner"`
}
```

**ツリー表現**:
```
TaskConfig (root)
├─ Version: int
├─ Task: TaskDetails
│  ├─ ID: string
│  ├─ Title: string
│  ├─ Repo: string
│  ├─ PRD: PRDDetails
│  │  ├─ Path: string
│  │  └─ Text: string
│  └─ Test: TestDetails
│     ├─ Command: string
│     └─ Cwd: string
└─ Runner: RunnerConfig
   ├─ Meta: MetaConfig
   │  ├─ Kind: string
   │  └─ Model: string
   └─ Worker: WorkerConfig
      ├─ Kind: string
      ├─ DockerImage: string
      ├─ MaxRunTimeSec: int
      └─ Env: map[string]string
```

## 各構造体の詳細

### 1. TaskConfig （ルート）

```go
type TaskConfig struct {
    Version int          `yaml:"version"`
    Task    TaskDetails  `yaml:"task"`
    Runner  RunnerConfig `yaml:"runner"`
}
```

**フィールド説明**:

| フィールド | 型 | 必須 | 説明 |
|----------|-----|------|------|
| `Version` | int | ○ | YAML スキーマバージョン（現在=1） |
| `Task` | TaskDetails | ○ | タスク詳細 |
| `Runner` | RunnerConfig | ○ |実行エンジン設定 |

**用途**: 全設定のルートコンテナ

**パース場所** (cmd/agent-runner/main.go):
```go
var cfg config.TaskConfig
yaml.Unmarshal(data, &cfg)
```

---

### 2. TaskDetails （タスク定義）

```go
type TaskDetails struct {
    ID    string      `yaml:"id"`
    Title string      `yaml:"title"`
    Repo  string      `yaml:"repo"`
    PRD   PRDDetails  `yaml:"prd"`
    Test  TestDetails `yaml:"test"`
}
```

**フィールド説明**:

| フィールド | 型 | 必須 | 説明 |
|----------|-----|------|------|
| `ID` | string | △ | タスク一意識別子（省略時は自動生成） |
| `Title` | string | △ | 人間が読めるタイトル |
| `Repo` | string | ○ | リポジトリパス（相対 or 絶対） |
| `PRD` | PRDDetails | ○ | 要件定義 |
| `Test` | TestDetails | △ | テスト設定（オプション） |

**YAML 例**:
```yaml
task:
  id: "TASK-001"          # オプション
  title: "ユーザー登録機能"
  repo: "."               # リポジトリ（Runner で絶対パスに変換）
  prd:
    text: |               # または path: "docs/prd.md"
      要件:
      - ユーザー登録フォーム実装
      - メール検証機能
  test:
    command: "npm test"
    cwd: "./"
```

---

### 3. PRDDetails （要件定義）

```go
type PRDDetails struct {
    Path string `yaml:"path"`
    Text string `yaml:"text"`
}
```

**設計パターン**: Path と Text は **相互排他的** （どちらか一方を使用）

| モード | 用途 | 利点 |
|--------|------|------|
| **Path** | `path: "docs/prd.md"` | 大規模PRD、ファイル管理 |
| **Text** | `text: \|` インラインPRD | 小規模タスク、YAML 内埋め込み |

**パース順序** (internal/core/runner.go:65-75):
```go
if r.Config.Task.PRD.Text != "" {
    taskCtx.PRDText = r.Config.Task.PRD.Text
} else if r.Config.Task.PRD.Path != "" {
    content, err := ioutil.ReadFile(r.Config.Task.PRD.Path)
    taskCtx.PRDText = string(content)
}
```

**動作**:
1. Text が指定 → Text を使用
2. Text が空 + Path が指定 → ファイルを読み込み
3. 両方空 → エラー（PRD not specified）

**設計理由**:
- 柔軟性：小規模タスクはインライン、大規模タスクはファイル参照
- 避けるべき形式：both 指定（Textが優先されて Path が無視される）

---

### 4. TestDetails （テスト設定）

```go
type TestDetails struct {
    Command string `yaml:"command"`
    Cwd     string `yaml:"cwd"`
}
```

**フィールド説明**:

| フィールド | 説明 |
|----------|------|
| `Command` | テスト実行コマンド（e.g., "npm test", "pytest", "go test ./..."） |
| `Cwd` | テスト実行ディレクトリ（相対パスはリポジトリ基準） |

**オプション性**: TaskDetails.Test フィールド全体が任意

**用途**: タスク完了後に自動テスト実行（将来実装予定）

---

### 5. RunnerConfig （実行エンジン設定）

```go
type RunnerConfig struct {
    Meta   MetaConfig   `yaml:"meta"`
    Worker WorkerConfig `yaml:"worker"`
}
```

**役割**: Meta-agentとWorkerの設定を保持

**パース先**:
- Meta → internal/meta/client.go:NewClient()
- Worker → internal/worker/executor.go:NewExecutor()

---

### 6. MetaConfig （Meta-agent設定）

```go
type MetaConfig struct {
    Kind  string `yaml:"kind"`
    Model string `yaml:"model"`
}
```

**フィールド説明**:

| フィールド | 値例 | 説明 |
|----------|------|------|
| `Kind` | "openai-chat" \| "mock" | LLM プロバイダー |
| `Model` | "gpt-5.2" | LLM モデル指定 |

**Kind 値の意味**:

| Kind | 動作 | テスト |
|------|------|--------|
| **"openai-chat"** | OpenAI API へ実際に接続 | ◯ 実運用、×テスト（API キー必須、料金発生） |
| **"mock"** | モック実装を使用、API呼び出しなし | ◯テスト、◯ローカル開発 |

**YAML 例**:
```yaml
runner:
  meta:
    kind: "mock"          # テスト時
    model: "gpt-5.2"      # 空の場合はビルトインデフォルトが使われる
```

**実装** (internal/meta/client.go):
```go
func NewClient(kind, apiKey, model, systemPrompt string) *Client {
    if model == "" {
        model = agenttools.DefaultMetaModel // デフォルト: gpt-5.2
    }
    // kind == "mock" の場合はシンプルな応答を返す
    // kind == "openai-chat" の場合は OpenAI API へ接続
}
```

---

### 7. WorkerConfig （Worker実行設定）

```go
type WorkerConfig struct {
    Kind          string            `yaml:"kind"`
    DockerImage   string            `yaml:"docker_image"`
    MaxRunTimeSec int               `yaml:"max_run_time_sec"`
    Env           map[string]string `yaml:"env"`
}
```

**フィールド説明**:

| フィールド | 型 | 説明 | 例 |
|----------|-----|------|-----|
| `Kind` | string | Worker タイプ（現在は "codex-cli" のみ） | "codex-cli" |
| `DockerImage` | string | Docker イメージ名 | "agent-runner-codex:latest" |
| `MaxRunTimeSec` | int | 実行タイムアウト（秒） | 1800（30分） |
| `Env` | map[string]string | 環境変数（Key=Value マップ） | `{ "CODEX_API_KEY": "env:..." }` |

**YAML 例**:
```yaml
runner:
  worker:
    kind: "codex-cli"
    docker_image: "agent-runner-codex:latest"
    max_run_time_sec: 1800
    env:
      CODEX_API_KEY: "env:CODEX_API_KEY"  # 環境変数から参照
      OPENAI_API_KEY: "sk-..."           # ハードコード（非推奨）
```

#### Env フィールドの環境変数参照

**パターン**: `"env:VARIABLE_NAME"` で ホスト環境変数から読み込み

```go
// 推奨
env:
  CODEX_API_KEY: "env:CODEX_API_KEY"

// 実行時に解決
envValue := os.Getenv("CODEX_API_KEY")
```

**理由**:
- Secrets を YAML に埋め込まない（セキュリティ）
- CI/CD パイプライン との連携が容易

---

## バージョニング戦略

### 現在：Version = 1

**スキーマ v1 の特徴**:
- TaskConfig, TaskDetails, PRDDetails, TestDetails 定義
- Meta と Worker の基本設定

### 将来：Version = 2 への移行例

**シナリオ**: 新フィールド `ValidationConfig` を追加

```yaml
version: 2  # <- バージョンアップ

task:
  id: "TASK-001"
  ...

runner:
  meta: ...
  worker: ...
  validation:          # <- 新規フィールド
    kind: "pytest"
    config_path: "pytest.ini"
```

**Go 構造体**:
```go
type RunnerConfig struct {
    Meta       MetaConfig       `yaml:"meta"`
    Worker     WorkerConfig     `yaml:"worker"`
    Validation *ValidationConfig `yaml:"validation,omitempty"`
}

type ValidationConfig struct {
    Kind       string `yaml:"kind"`
    ConfigPath string `yaml:"config_path"`
}
```

**互換性管理**:
```go
// main.go のパーサー
var cfg config.TaskConfig
yaml.Unmarshal(data, &cfg)

if cfg.Version == 1 {
    // v1 パーサロジック
} else if cfg.Version == 2 {
    // v2 パーサロジック（Validation 処理追加）
}
```

**推奨事項**:
- v1 から v2 への移行：段階的（v2 で v1 互換性保持）
- Breaking change（互換性破棄） は Major version で明示的に

---

## デフォルト値と最小構成

### 最小限の YAML（必須フィールドのみ）

```yaml
version: 1

task:
  repo: "."
  prd:
    text: |
      Implement basic feature

runner:
  meta:
    kind: "mock"
  worker:
    kind: "codex-cli"
```

**省略時のデフォルト**:
- `task.id`: 自動生成（現在は Runner が生成）
- `task.title`: 空文字列
- `meta.model`: "gpt-4-turbo"
- `worker.docker_image`: "ghcr.io/biwakonbu/agent-runner-codex:latest"
- `worker.max_run_time_sec`: デフォルト設定なし（Commander で管理）

---

## 拡張ガイド

### 新しいフィールドを追加する場合

**例: DeploymentConfig を追加**

1. **構造体定義** (config.go)
   ```go
   type DeploymentConfig struct {
       Target    string `yaml:"target"`     // "staging", "production"
       Strategy  string `yaml:"strategy"`   // "blue-green", "canary"
   }

   type RunnerConfig struct {
       Meta       MetaConfig       `yaml:"meta"`
       Worker     WorkerConfig     `yaml:"worker"`
       Deployment *DeploymentConfig `yaml:"deployment,omitempty"`
   }
   ```

2. **YAML スキーマドキュメント更新** (このファイル)
   - DeploymentConfig セクション追加

3. **テスト追加** (config_test.go)
   ```go
   func TestUnmarshal_WithDeploymentConfig(t *testing.T) {
       yaml := `version: 1
   runner:
     deployment:
       target: "staging"
       strategy: "blue-green"
   `
       // パース検証
   }
   ```

4. **Runner 実装更新** (internal/core/runner.go)
   - DeploymentConfig を活用するロジック追加

---

## YAML パース・エラーハンドリング

### パース場所

**main.go** (cmd/agent-runner/main.go):
```go
var cfg config.TaskConfig
if err := yaml.Unmarshal(stdin_data, &cfg); err != nil {
    log.Fatalf("YAML parse error: %v", err)
}
```

### 一般的なエラーと対応

| エラー | 原因 | 対応 |
|--------|------|------|
| `yaml: unmarshal errors...` | フィールド型不一致 | YAML の値の型を確認（例：string vs int） |
| `yaml: field version not found` | フィールド名誤り | yaml tag のスペル確認 |
| `PRD not specified` | PRD.Path と PRD.Text 両方空 | どちらか一方を指定 |
| `failed to read PRD file` | PRD.Path が無効 | ファイルパスを絶対パスで指定 |

---

## ベストプラクティス

### 1. 環境変数の活用

```yaml
# ◯ 推奨：環境変数参照
env:
  CODEX_API_KEY: "env:CODEX_API_KEY"
  OPENAI_API_KEY: "env:OPENAI_API_KEY"

# ✗ 非推奨：ハードコード
env:
  CODEX_API_KEY: "sk-xxx"  # キーが露出
```

### 2. リポジトリパスの指定

```yaml
# ◯ 推奨（内部で絶対パスに変換）
repo: "."

# ◯ 推奨（絶対パス指定）
repo: "/Users/dev/projects/my-repo"

# ✗ 相対パスの複雑な指定
repo: "../../parent-dir/repo"  # 予測困難
```

### 3. PRD 指定方式の選択

```yaml
# ◯ 小規模タスク（インライン）
prd:
  text: |
    要件:
    - Feature X

# ◯ 大規模タスク（ファイル参照）
prd:
  path: "docs/prd.md"

# ✗ 両方指定（Text が優先される）
prd:
  path: "docs/prd.md"
  text: "override"  # 無視される
```

### 4. Kind/Model の組み合わせ

```yaml
# テスト・開発時
runner:
  meta:
    kind: "mock"

# 本番運用
runner:
  meta:
    kind: "openai-chat"
    model: "gpt-4-turbo"
```

---

## サンプル設定ファイル

### Go プロジェクト用（sample_task_go.yaml）

```yaml
version: 1

task:
  id: "TASK-001"
  title: "Go API 実装"
  repo: "."
  prd:
    text: |
      要件：
      - RESTful API エンドポイント実装
      - テストカバレッジ 80% 以上
  test:
    command: "go test ./..."
    cwd: "./"

runner:
  meta:
    kind: "mock"
    model: "gpt-4-turbo"
  worker:
    kind: "codex-cli"
    docker_image: "agent-runner-codex:latest"
    max_run_time_sec: 1800
    env:
      CODEX_API_KEY: "env:CODEX_API_KEY"
```

### Codex 統合テスト用（test_codex_task.yaml）

```yaml
version: 1

task:
  id: "CODEX-TEST"
  title: "Codex 統合テスト"
  repo: "."
  prd:
    path: "docs/test_requirements.md"

runner:
  meta:
    kind: "mock"
  worker:
    kind: "codex-cli"
    docker_image: "agent-runner-codex:latest"
```

---

## 関連ドキュメント

- `/CLAUDE.md` - タスク YAML 形式の概要
- [core/CLAUDE.md](../core/CLAUDE.md) - TaskConfig の使用箇所
- [meta/CLAUDE.md](../meta/CLAUDE.md) - MetaConfig の処理
- [worker/CLAUDE.md](../worker/CLAUDE.md) - WorkerConfig の処理
- `/sample_task_go.yaml` - Go プロジェクト用サンプル
- `/test_codex_task.yaml` - Codex 統合テスト用サンプル
