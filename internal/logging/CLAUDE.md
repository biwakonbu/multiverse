# logging/CLAUDE.md - 統一構造化ロギング

このパッケージは multiverse システム全体で使用する統一された構造化ロギング機能を提供します。

## パッケージ概要

`logging` パッケージは、Trace ID 伝播、複数ログレベル、JSON/Text フォーマット出力をサポートする構造化ロギングを提供します。Go 標準の `log/slog` パッケージをベースに構築されています。

## 主要概念

### Config

ロガー設定を保持する構造体。3 つのプリセットが用意されています。

```go
// デフォルト設定（開発用）
DefaultConfig()    // Level: INFO, JSONFormat: false, AddSource: false

// 本番設定
ProductionConfig() // Level: INFO, JSONFormat: true, AddSource: true

// デバッグ設定
DebugConfig()      // Level: DEBUG, JSONFormat: false, AddSource: true
```

### Trace ID 伝播

リクエスト単位でログを追跡するための Trace ID をコンテキスト経由で伝播します。

```go
// Trace ID をコンテキストに設定
ctx := logging.ContextWithTraceID(ctx, "trace-abc-123")

// コンテキストから Trace ID を取得
traceID := logging.TraceIDFromContext(ctx)

// ロガーに Trace ID を付与
logger := logging.WithTraceID(logger, ctx)
```

### ログコンテキスト構造体

特定のドメインに関するログ属性を一括で追加するためのヘルパー構造体。

| 構造体 | 用途 | 属性 |
|-------|------|------|
| `LogRequest` | API リクエストログ | Method, URL, StatusCode, DurationMs, RequestSize, ResponseSize, Error |
| `TaskLogContext` | タスク関連ログ | TaskID, Title, State, LoopCount |
| `WorkerLogContext` | Worker 関連ログ | ContainerID, Image, Command, ExitCode, DurationMs |

## 実装パターン

### ロガー初期化

```go
// 環境に応じた設定でロガーを作成
var cfg logging.Config
if os.Getenv("ENV") == "production" {
    cfg = logging.ProductionConfig()
} else {
    cfg = logging.DefaultConfig()
}
logger := logging.NewLogger(cfg)
```

### コンポーネント別ロガー

```go
// コンポーネント名を付与したロガーを作成
metaLogger := logging.WithComponent(logger, "meta-client")
workerLogger := logging.WithComponent(logger, "worker-executor")
```

### 実行時間計測

```go
start := time.Now()
// ... 処理 ...
logger.Info("処理完了", logging.LogDuration(start))
```

### API リクエストログ

```go
reqLog := logging.LogRequest{
    Method:       "POST",
    URL:          "https://api.openai.com/v1/chat/completions",
    StatusCode:   200,
    DurationMs:   1234.5,
    RequestSize:  500,
    ResponseSize: 2000,
}
logger.Info("API request completed", reqLog.ToAttrs()...)
```

## 定数

| 定数 | 値 | 用途 |
|------|---|------|
| `TraceIDKey` | `"trace_id"` | コンテキストキー |
| `AttrTraceID` | `"trace_id"` | ログ属性名 |
| `AttrComponent` | `"component"` | コンポーネント属性名 |
| `AttrDuration` | `"duration_ms"` | 実行時間属性名 |

## テスト戦略

### ユニットテスト

- `logging_test.go` で全機能をカバー
- バッファを使用してログ出力を検証
- JSON フォーマットの妥当性を検証

### テスト実行

```bash
go test ./internal/logging/...
```

## 拡張・カスタマイズ

### 新しいログコンテキスト追加

1. 構造体を定義
2. `ToAttrs() []slog.Attr` メソッドを実装

```go
type CustomLogContext struct {
    Field1 string
    Field2 int
}

func (c CustomLogContext) ToAttrs() []slog.Attr {
    return []slog.Attr{
        slog.String("field1", c.Field1),
        slog.Int("field2", c.Field2),
    }
}
```

### カスタムハンドラー

`slog.Handler` インターフェースを実装することで、独自の出力先やフォーマットを追加可能です。

## 関連ドキュメント

- [../../docs/specifications/logging-specification.md](../../docs/specifications/logging-specification.md): ロギング仕様書
- [../core/CLAUDE.md](../core/CLAUDE.md): Core パッケージ（ログ使用例）
- [../meta/CLAUDE.md](../meta/CLAUDE.md): Meta パッケージ（API リクエストログ）
- [../worker/CLAUDE.md](../worker/CLAUDE.md): Worker パッケージ（コンテナログ）
