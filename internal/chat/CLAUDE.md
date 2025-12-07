# chat パッケージ

## 責務

チャットインターフェースを通じたユーザー入力の処理とタスク生成を担当する。
v2.0 のコア機能「チャット駆動」を実現する中心パッケージ。

## 主要コンポーネント

### ChatHandler (`handler.go`)

ユーザーメッセージを受け取り、Meta-agent を呼び出してタスクを生成する。

```go
type Handler struct {
    Meta         MetaClient              // Meta-agent クライアント
    TaskStore    *orchestrator.TaskStore // タスク永続化
    SessionStore *ChatSessionStore       // セッション永続化
    WorkspaceID  string                  // ワークスペースID
    ProjectRoot  string                  // プロジェクトルートパス
    metaTimeout  time.Duration           // Meta-agent 呼び出しタイムアウト
}
```

**主要メソッド:**
- `HandleMessage(ctx, sessionID, message)` - メッセージ処理とタスク生成
- `CreateSession(ctx)` - 新規セッション作成
- `GetHistory(ctx, sessionID)` - 履歴取得
- `SetMetaTimeout(timeout)` - タイムアウト設定

### タイムアウト設定

```go
// Meta-agent 呼び出しのデフォルトタイムアウト（15分）
const DefaultChatMetaTimeout = 15 * time.Minute
```

LLM によるタスク分解は時間がかかるため、十分な時間を確保する。
`SetMetaTimeout()` メソッドで個別に設定可能。

### ChatSessionStore (`session_store.go`)

チャットセッションとメッセージの JSONL 永続化を管理。

**ファイル構造:**
```
~/.multiverse/workspaces/<id>/chat/
├── <session-id>.meta.json   # セッションメタデータ
└── <session-id>.jsonl       # メッセージ履歴（1行1メッセージ）
```

## 処理フロー

```
ユーザー入力
    ↓
HandleMessage()
    ├── ユーザーメッセージ保存
    ├── コンテキスト収集（既存タスク、会話履歴）
    ├── Meta-agent.Decompose() 呼び出し
    ├── タスク永続化（一時ID→正式ID変換）
    └── アシスタント応答保存
    ↓
ChatResponse（生成タスク、理解、コンフリクト情報）
```

## 依存関係

- `internal/meta` - Meta-agent との通信
- `internal/orchestrator` - Task の永続化
- `internal/logging` - 構造化ログ

## テスト戦略

- モック MetaClient を使用したユニットテスト
- SessionStore の JSONL 読み書きテスト
- 一時ID→正式ID変換のテスト

## 設計原則

- **インターフェース依存**: MetaClient はインターフェースとして定義し、モック可能
- **冪等性**: 同一メッセージの再処理でも整合性を維持
- **エラー耐性**: Meta-agent エラー時もユーザーに応答を返す
