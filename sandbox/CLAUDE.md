# sandbox/CLAUDE.md - Worker Docker イメージ定義

このディレクトリは Worker agent が実行される Docker サンドボックス環境のイメージ定義を管理します。

## ディレクトリ概要

`sandbox/` ディレクトリには、AgentRunner の Worker が安全に実行されるための Docker イメージ定義が含まれています。このイメージは開発タスクを隔離された環境で実行するためのランタイムを提供します。

## ファイル構成

| ファイル | 役割 |
|---------|------|
| Dockerfile | Worker ランタイムイメージ定義 |

## Dockerfile 構成

### ベースイメージ

```dockerfile
FROM python:3.11-slim
```

Python 3.11 の軽量イメージをベースに使用。データサイエンス・スクリプティング用途に適しています。

### インストールされるツール

| カテゴリ | ツール | 用途 |
|---------|-------|------|
| バージョン管理 | git | リポジトリ操作 |
| HTTP クライアント | curl | API 通信・ファイルダウンロード |
| エディタ | vim | ファイル編集 |
| JavaScript ランタイム | nodejs, npm | Node.js プロジェクト・Codex CLI |
| JSON 処理 | jq | JSON パース・加工 |
| AI コーディング | @openai/codex | Codex CLI（Worker 実行） |

### ディレクトリ構造

```
/root/.codex/     # Codex CLI 設定・認証
/workspace/       # 作業ディレクトリ（プロジェクトマウントポイント）
```

## ビルド方法

```bash
# イメージビルド
docker build -t agent-runner-codex:latest sandbox/

# ビルド確認
docker images | grep agent-runner-codex
```

## 使用方法

AgentRunner Core の Worker 設定で Docker イメージを指定します。

### タスク YAML 設定例

```yaml
runner:
  worker:
    kind: "codex-cli"
    docker_image: "agent-runner-codex:latest"
    max_run_time_sec: 1800
    env:
      CODEX_API_KEY: "env:CODEX_API_KEY"
```

### 手動実行（デバッグ用）

```bash
# コンテナ起動
docker run -it --rm \
  -v $(pwd):/workspace/project \
  -e CODEX_API_KEY=$CODEX_API_KEY \
  agent-runner-codex:latest \
  /bin/bash

# Codex CLI 動作確認
codex --version
```

## マウント仕様

Worker 実行時、以下のマウントが行われます：

| ホスト | コンテナ | モード | 用途 |
|-------|---------|--------|------|
| プロジェクトルート | `/workspace/project` | rw | 開発対象リポジトリ |
| Codex 認証ファイル | `/root/.codex/auth.json` | ro | API 認証（オプション） |

## 拡張ガイドライン

### 新しいツール追加

1. Dockerfile の `apt-get install` セクションに追加
2. イメージを再ビルド
3. `worker-interface.md` 仕様を更新

```dockerfile
RUN apt-get update && apt-get install -y \
    git \
    curl \
    vim \
    nodejs \
    npm \
    jq \
    新しいツール \    # 追加
    && rm -rf /var/lib/apt/lists/*
```

### 別の Worker ランタイム追加

新しい Worker タイプ（例: Claude CLI）をサポートする場合：

1. 新しい Dockerfile を作成（例: `Dockerfile.claude`）
2. 対応するイメージをビルド
3. `internal/worker/executor.go` に新しい Worker kind を追加

### イメージサイズ最適化

- 不要なキャッシュは `rm -rf /var/lib/apt/lists/*` で削除済み
- マルチステージビルドは現時点で不要（シンプルな構成のため）
- 必要に応じて `.dockerignore` を追加

## セキュリティ考慮事項

### 実行時の隔離

- 各タスクは独立したコンテナで実行
- ネットワークアクセスはデフォルトで有効（API 通信のため）
- ファイルシステムはマウントされたディレクトリに限定

### 認証情報の取り扱い

- API キーは環境変数経由で注入
- 認証ファイルは読み取り専用でマウント
- コンテナ終了時に一時データは破棄

## 関連ドキュメント

- [../internal/worker/CLAUDE.md](../internal/worker/CLAUDE.md): Worker 実行・サンドボックス管理
- [../docs/specifications/worker-interface.md](../docs/specifications/worker-interface.md): Worker インターフェース仕様
- [../test/CLAUDE.md](../test/CLAUDE.md): Docker テスト（-tags=docker）
