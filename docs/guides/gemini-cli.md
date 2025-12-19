# Gemini CLI 統合ガイド

このドキュメントでは、Google の Gemini CLI を multiverse で使用するための設定と運用ノウハウをまとめています。

## 概要

Gemini CLI は Google が提供するオープンソースの AI エージェントで、ターミナルから直接 Gemini モデルにアクセスできます。

- **公式リポジトリ**: https://github.com/google-gemini/gemini-cli
- **ドキュメント**: https://geminicli.com/docs/

### 主な特徴

- 無料枠: 60 リクエスト/分、1,000 リクエスト/日（個人 Google アカウント）
- 1M トークンのコンテキストウィンドウ
- 組み込みツール: Google 検索、ファイル操作、シェルコマンド、Web フェッチ
- MCP（Model Context Protocol）サポート

## 利用可能なモデル

### 推奨モデル

| モデル ID | 特徴 | 用途 |
|-----------|------|------|
| `gemini-3-flash-preview` | 最新のマルチモーダル、低レイテンシ | **デフォルト・日常的なタスク** |
| `gemini-3-pro-preview` | 最新のマルチモーダル、1M入力/65k出力 | 高度なタスク |
| `gemini-2.5-pro` | 高度な推論、STEM 分析、安定版 | 複雑なコード生成・分析 |
| `gemini-2.5-flash` | 価格・性能バランス、安定版 | 日常的な開発作業 |
| `gemini-2.5-flash-lite` | 超高速・低コスト | 大量リクエスト処理 |

### プレビューモデル

| モデル ID | 特徴 | 注意事項 |
|-----------|------|----------|
| `gemini-3-flash-preview` | 最新のマルチモーダル（**デフォルト**） | 2週間前通知で変更の可能性 |
| `gemini-3-pro-preview` | 最新のマルチモーダル | プレビュー版 |
| `gemini-2.5-flash-preview-09-2025` | Flash のプレビュー版 | プレビュー版 |

### 特殊モデル

| モデル ID | 用途 |
|-----------|------|
| `gemini-2.5-flash-preview-tts` | テキスト読み上げ |
| `gemini-2.5-flash-image` | 画像生成 |
| `gemini-2.5-flash-native-audio-preview-09-2025` | ライブオーディオ |

## 環境設定

### 認証設定

```bash
# 方法 1: 環境変数（推奨）
export GEMINI_API_KEY="your-api-key"

# 方法 2: Google Cloud 認証
export GOOGLE_API_KEY="your-api-key"

# 方法 3: Vertex AI 経由
export GOOGLE_GENAI_USE_VERTEXAI=true
export GOOGLE_CLOUD_PROJECT="your-project-id"
```

### .env ファイル

`~/.gemini/.env`（グローバル）または `./.gemini/.env`（プロジェクト）に設定可能:

```bash
GEMINI_API_KEY=your-api-key
GEMINI_MODEL=gemini-3-flash-preview
```

## CLI オプション

### 基本コマンド

```bash
# インタラクティブモード
gemini

# 非インタラクティブ（プロンプトモード）
gemini -p "コードをレビューして"

# モデル指定
gemini -m gemini-3-pro-preview

# JSON 出力
gemini -p "質問" --output-format json

# 複数ディレクトリをコンテキストに追加
gemini --include-directories ../lib,../docs
```

### 主要フラグ

| フラグ | 説明 |
|--------|------|
| `-m`, `--model` | 使用するモデルを指定 |
| `-p` | プロンプトモード（非インタラクティブ） |
| `--output-format` | 出力形式（`json`, `stream-json`） |
| `--include-directories` | コンテキストに含めるディレクトリ |
| `--yolo` | ツール呼び出しを自動承認 |

## 設定ファイル（settings.json）

### 設定の優先順位

1. コマンドライン引数（最優先）
2. 環境変数・.env ファイル
3. システム設定（`/etc/gemini-cli/settings.json`）
4. プロジェクト設定（`.gemini/settings.json`）
5. ユーザー設定（`~/.gemini/settings.json`）
6. デフォルト値（最低優先）

### 推奨設定

```json
{
  "theme": "Default",
  "vimMode": false,
  "hideBanner": true,
  "autoAccept": false,
  "coreTools": ["read_file", "write_file", "run_shell_command"],
  "sandbox": false,
  "checkpointing": {
    "enabled": true
  },
  "summarizeToolOutput": {
    "run_shell_command": {
      "enabled": true,
      "tokenBudget": 2000
    }
  }
}
```

### 主要設定項目

#### コンテキスト設定

| 設定 | 型 | 説明 |
|------|-----|------|
| `contextFileName` | string/array | コンテキストファイル名（デフォルト: `GEMINI.md`） |
| `includeDirectories` | array | コンテキストに含めるディレクトリ |
| `loadMemoryFromIncludeDirectories` | boolean | 含めたディレクトリから GEMINI.md を読み込む |

#### ツール設定

| 設定 | 型 | 説明 |
|------|-----|------|
| `coreTools` | array | 有効にするツール |
| `excludeTools` | array | 除外するツール |
| `autoAccept` | boolean | 安全なツール実行を自動承認 |

#### MCP サーバー設定

```json
{
  "mcpServers": {
    "my-server": {
      "command": "node",
      "args": ["server.js"],
      "env": {},
      "timeout": 30000
    }
  }
}
```

#### サンドボックス設定

| 設定 | 型 | 説明 |
|------|-----|------|
| `sandbox` | boolean/string | サンドボックス有効化（`true`, `"docker"`, `"podman"`） |

## GEMINI.md（コンテキストファイル）

プロジェクトの説明をモデルに提供するファイル。

### 配置場所と優先順位

1. `~/.gemini/GEMINI.md` - グローバル設定
2. プロジェクトルートから現在ディレクトリまでの祖先
3. サブディレクトリの GEMINI.md

### 初期化

```bash
gemini /init
```

### 推奨構成

```markdown
# プロジェクト名

## 概要
プロジェクトの目的と主要機能

## 技術スタック
- 言語: Go 1.23
- フレームワーク: ...

## ディレクトリ構造
- `cmd/` - エントリポイント
- `internal/` - 内部パッケージ

## コーディング規約
- コメントは日本語
- 変数名は英語

## よく使うコマンド
- `go test ./...` - テスト実行
- `go build ./cmd/...` - ビルド
```

## multiverse での設定

### タスク YAML 設定

```yaml
runner:
  worker:
    kind: "gemini-cli"
    model: "gemini-3-flash-preview"  # または gemini-3-pro-preview
    max_run_time_sec: 300
    env:
      GEMINI_API_KEY: "env:GEMINI_API_KEY"
```

### ProviderConfig

```go
cfg := agenttools.ProviderConfig{
    CLIPath:  "gemini",
    Model:    "gemini-3-flash-preview",
    ExtraEnv: map[string]string{
        "GEMINI_API_KEY": os.Getenv("GEMINI_API_KEY"),
    },
    Flags: []string{},
}
provider := agenttools.NewGeminiProvider(cfg)
```

## 運用ノウハウ

### モデル選択の指針

| シナリオ | 推奨モデル | 理由 |
|----------|-----------|------|
| 日常的なタスク・デフォルト | `gemini-3-flash-preview` | 低レイテンシ・最新世代 |
| 高度なタスク | `gemini-3-pro-preview` | 最新のマルチモーダル能力 |
| 安定性重視のコード生成 | `gemini-2.5-pro` | 高度な推論能力・安定版 |
| 日常的なコード生成 | `gemini-2.5-flash` | バランスが良く安定・低コスト |
| 大量のファイル処理 | `gemini-2.5-flash-lite` | 低コスト・高速 |

### レート制限対策

無料枠の制限（60 req/min、1,000 req/day）に注意:

```go
// リトライロジックの実装例
func withRetry(fn func() error, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        if isRateLimitError(err) {
            time.Sleep(time.Duration(i+1) * time.Second)
            continue
        }
        return err
    }
    return fmt.Errorf("max retries exceeded")
}
```

### コンテキスト最適化

1M トークンのコンテキストを効率的に使用:

```json
{
  "fileFiltering": {
    "respectGitIgnore": true,
    "enableRecursiveFileSearch": true
  },
  "summarizeToolOutput": {
    "run_shell_command": {
      "enabled": true,
      "tokenBudget": 2000
    }
  }
}
```

### トラブルシューティング

#### 認証エラー

```
Error: API key not found
```

**対策**:
1. `GEMINI_API_KEY` 環境変数を確認
2. `~/.gemini/.env` ファイルを確認
3. API キーの有効性を確認

#### モデルが見つからない

```
Error: Model not found: gemini-3-pro
```

**対策**:
- 正しいモデル ID を使用（`gemini-3-pro-preview` など）
- プレビューモデルは変更される可能性があることを認識

#### タイムアウト

```
Error: Request timeout
```

**対策**:
1. `max_run_time_sec` を増やす
2. プロンプトを簡潔にする
3. `gemini-2.5-flash-lite` で高速化

## 参考リンク

- [Gemini CLI GitHub](https://github.com/google-gemini/gemini-cli)
- [Gemini API モデル一覧](https://ai.google.dev/gemini-api/docs/models)
- [Gemini CLI 設定ドキュメント](https://github.com/google-gemini/gemini-cli/blob/main/docs/cli/configuration.md)
- [Google Codelabs - Gemini CLI ハンズオン](https://codelabs.developers.google.com/gemini-cli-hands-on)
