# Codex Integration Test

このディレクトリには、実際の Codex CLI を使用した統合テストが含まれています。

## 前提条件

1. **Codex 認証の設定**

   - ホストマシンに `~/.codex/auth.json` が存在する必要があります
   - Codex CLI は認証情報を Docker コンテナにマウントして使用します

2. **Docker イメージのビルド**
   ```bash
   docker build -t agent-runner-codex:latest sandbox/
   ```

## テストの実行

### 方法 1: テストスクリプトを使用

```bash
./run_codex_test.sh
```

### 方法 2: 直接実行

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
