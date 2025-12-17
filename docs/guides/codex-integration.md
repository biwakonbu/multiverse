# Codex Integration Test

このディレクトリには、実際の Codex CLI を使用した統合テストが含まれています。

## モデル（参照 URL）

- https://platform.openai.com/docs/pricing

このプロジェクトのデフォルトは `gpt-5.2`（Meta-agent）と `gpt-5.1-codex`（Worker）です（実装: `internal/agenttools/codex.go`）。
必要に応じて `gpt-5.1-codex-mini`（ショートハンド: `5.1-codex-mini`）も利用できます（実装: `internal/agenttools/openai_models.go`）。

## 前提条件

1. **Codex 認証の設定**

   - ホストマシンに `~/.codex/auth.json` が存在する必要があります
   - Codex CLI は認証情報を Docker コンテナにマウントして使用します

2. **Docker イメージのビルド**
   ```bash
   docker build -t agent-runner-codex:latest sandbox/
   ```

## テストの実行

### 方法 1: go test で実行（推奨）

```bash
# Codex テストのみ
go test -tags=codex -timeout=10m ./test/codex/...

# 詳細表示
go test -v -tags=codex -timeout=10m ./test/codex/...
```

### 方法 2: テストスクリプトを使用

```bash
./run_codex_test.sh
```

### 方法 3: 直接実行

```bash
go run cmd/agent-runner/main.go < test_codex_task.yaml
```

## テスト内容

`test_codex_task.yaml` は以下をテストします：

- 簡単な電卓プログラム（calculator.py）の作成
- Codex CLI が Docker サンドボックス内で正しく動作すること
- ファイルがリポジトリに正しく保存されること

## 結果の確認

テスト実行後、以下を確認してください：

1. `.agent-runner/task-TASK-CODEX-TEST.md` - タスクノート
2. `calculator.py` - Codex が生成したファイル（リポジトリルートに作成されるはず）

## トラブルシューティング

### Codex 認証エラー

```
Error: Codex authentication failed
```

→ `~/.codex/auth.json` が存在し、有効な認証情報が含まれていることを確認してください。

### Docker コンテナ起動エラー

```
Error: failed to start sandbox
```

→ Docker デーモンが起動していることを確認してください。

### signal: killed エラー

```
タスク分解に失敗しました: codex CLI call failed: codex CLI 呼び出し失敗: signal: killed
```

**原因**: タイムアウトによりプロセスが強制終了されました。

**対策**:

1. **タイムアウト設定の確認**:
   - ChatHandler: デフォルト 15 分（`DefaultChatMetaTimeout`）
   - Meta-agent: デフォルト 10 分（`DefaultMetaAgentTimeout`）

2. **ログの確認**:
   - プロセスがどの段階でタイムアウトしたかを確認
   - ネットワーク遅延や API レート制限の可能性をチェック

3. **タイムアウト延長**（必要な場合）:
   ```go
   // chat/handler.go
   handler.SetMetaTimeout(20 * time.Minute)
   ```

### YAML パースエラー

```
failed to parse YAML response: mapping values are not allowed in this context
```

**原因**: Codex CLI の出力にヘッダー情報が含まれており、YAML パーサーがそれを解釈できませんでした。

**対策**:

1. `extractYAML()` 関数が正しく YAML 部分を抽出しているか確認
2. Codex CLI の出力形式が変わっていないか確認

**Codex CLI の出力形式**:

```
OpenAI Codex v0.65.0 (research preview)
--------
workdir: /path/to/project
model: gpt-5.2
provider: openai
--------
user
プロンプト内容...
codex
type: decompose
version: 1
payload:
  understanding: "..."
```

`extractYAML()` は `type:` で始まる行から YAML を抽出します。
