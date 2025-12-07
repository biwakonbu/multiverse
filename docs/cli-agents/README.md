# CLI エージェントナレッジ

このディレクトリには、multiverse IDE で使用する CLI エージェントツールのナレッジを管理します。

## ディレクトリ構造

```
docs/cli-agents/
├── README.md           # このファイル
├── codex/              # Codex CLI
│   ├── CLAUDE.md       # AI 向けナレッジ
│   └── version-X.X.X.md # バージョン固有仕様
├── gemini/             # Gemini CLI（将来拡張）
├── claude-code/        # Claude Code（将来拡張）
└── cursor/             # Cursor CLI（将来拡張）
```

## 共通原則

### サンドボックス方針

**全ての CLI エージェントツールは Docker コンテナ内で実行され、CLI 内部のサンドボックスは無効化される。**

詳細は [サンドボックス方針](../design/sandbox-policy.md) を参照。

### ナレッジ管理ルール

1. **CLAUDE.md**: AI（Claude）が参照するための構造化されたナレッジ
   - 現在対応しているバージョン
   - 必須フラグと設定
   - デフォルト値
   - 使用例

2. **version-X.X.X.md**: バージョン固有の詳細仕様
   - そのバージョンで利用可能なフラグ一覧
   - 前バージョンからの変更点
   - 既知の問題

### 更新タイミング

- CLI ツールのバージョンアップ時
- 新しいフラグ・機能の追加時
- 問題発生時の調査結果

## 対応 CLI ツール

| CLI ツール | ステータス | 対応バージョン |
|-----------|----------|---------------|
| Codex CLI | ✅ 対応済み | 0.65.0 |
| Gemini CLI | ⏳ 未対応 | - |
| Claude Code | ⏳ 未対応 | - |
| Cursor CLI | ⏳ 未対応 | - |

## 関連ドキュメント

- [サンドボックス方針](../design/sandbox-policy.md)
- [AgentToolProvider 設計](../design/architecture.md#agenttoolprovider-設計phase-4-拡張)
- [Worker インターフェース仕様](../specifications/worker-interface.md)
