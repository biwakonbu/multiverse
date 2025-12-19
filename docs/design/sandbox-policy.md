# サンドボックス方針

最終更新: 2025-12-17

## 基本原則（絶対遵守）

Docker コンテナが外部サンドボックスとして機能するため、CLI エージェントツール内部のサンドボックス機能は**無効化**し、最大限の権限を与える。

**この方針は multiverse IDE の設計思想の根幹であり、全ての CLI エージェントツールに適用される絶対ルールである。**

## 理由

1. **Docker コンテナ自体が隔離環境として十分な保護を提供**
   - ファイルシステムの隔離
   - プロセスの隔離
   - ネットワークの制御

2. **二重サンドボックスの問題回避**
   - CLI ツール内部で二重にサンドボックスを有効にすると、ファイル操作・コマンド実行に不必要な制限がかかる
   - タスク実行に必要な権限が不足し、作業が失敗する

3. **自律実行の要件**
   - Worker エージェントはコード編集、テスト実行、ビルドなど多様な操作を行う
   - これらの操作には十分な権限が必要

## 全 CLI ツール共通設定

| CLI ツール | 無効化フラグ | 備考 |
|-----------|-------------|------|
| Codex CLI | `--dangerously-bypass-approvals-and-sandbox` | 0.65.0 で確認 |
| Gemini CLI | （該当フラグなし: `--sandbox` を使用しない） | `--yolo` で承認を自動化 |
| Claude Code | （該当フラグなし: 現実装は `-p`） | `internal/agenttools/claude.go` |
| Cursor CLI | （TBD: 実装時に調査） | |

## 安全性の保証

### Docker コンテナによる保護

- **ファイルシステム隔離**: コンテナ内のファイルシステムはホストから隔離
- **マウント制御**: ホストファイルシステムへのアクセスは明示的なマウント設定でのみ許可
- **ネットワーク制御**: Docker ネットワーク設定でネットワークアクセスを制御
- **リソース制限**: CPU・メモリ・ディスクの使用量を制限可能

### マウント設定

```yaml
# 推奨マウント設定
volumes:
  - type: bind
    source: ${PROJECT_ROOT}
    target: /workspace/project
    # read-write（作業用）

  - type: bind
    source: ~/.codex/auth.json
    target: /root/.codex/auth.json
    read_only: true  # 認証情報は読み取り専用

  - type: bind
    source: ~/.config/claude
    target: /root/.config/claude
    read_only: true  # 認証情報は読み取り専用
```

### ネットワーク設定

```yaml
# 推奨ネットワーク設定
networks:
  - agent-network  # 必要に応じて外部アクセスを許可
```

## 実装ガイドライン

### AgentToolProvider 実装時の必須事項

1. **サンドボックス無効化フラグを必ず指定する**
   - Docker 内実行であることを前提とし、CLI 内部のサンドボックスを無効化

2. **承認プロンプトを無効化する**
   - 自律実行のため、ユーザー確認なしで操作を実行

3. **フルアクセス権限を付与する**
   - ファイル操作、コマンド実行に必要な全権限を付与

### 禁止事項

1. **ホストで直接 CLI を実行しない**
   - 必ず Docker コンテナ内で実行すること
   - ホストで `--dangerously-bypass-approvals-and-sandbox` を使用してはならない

2. **サンドボックスを有効にしたまま Docker 内で実行しない**
   - 二重サンドボックスは問題を引き起こす

## 関連ドキュメント

- [Worker インターフェース仕様](../specifications/worker-interface.md)
- [CLI エージェントナレッジ](../cli-agents/README.md)
- [Codex CLI ナレッジ](../cli-agents/codex/CLAUDE.md)
