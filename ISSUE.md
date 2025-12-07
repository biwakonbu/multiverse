# ISSUE Log (2025-12-07)

## Open Items

- [ ] Meta 層が CLI サブスクリプション未対応（API キー不要方針と不整合）
  - `app.go` の `newMetaClientFromEnv()` と `chat.NewHandler` が HTTP クライアント (OPENAI_API_KEY 前提) を生成し、LLMConfigStore/設定画面を参照しない。Codex / Gemini / Claude Code / Cursor などの CLI セッション再利用方針と乖離。
  - 対応: Meta プロバイダを CLI 実装に差し替える抽象化を入れ、LLMConfigStore 経由でプロバイダを切替できるようにする。設定変更後にチャットハンドラを再初期化する経路も必要。

- [ ] LLM 設定 UI が実行系に反映されず、API キー前提の表示が残存
  - `LLMSettings` は Kind/Model を保存するが、CLI セッション状態を表示できず、`TestLLMConnection` も OpenAI HTTP 前提で CLI セッションを検証できない。API キーは不要なので UI をセッション表示に置換する必要。
  - 対応: CLI セッション検証用のテストエンドポイントに差し替え、設定保存を Meta 層の初期化に反映。API キー保存/表示を廃止し、セッション検知を表示。

- [ ] 実行ログのリアルタイム配信が未実装
  - `internal/orchestrator/executor.go` は `CombinedOutput` を一括取得するのみで、Wails Events によるストリーミングが無い。`TaskLogView` 相当の UI も未接続のため AC-P4-05 を満たさない。
  - 対応: stdout/stderr を逐次読み取り、`task:log` イベントを送出する実装と UI 連携を追加。

- [ ] Codex CLI セッションの存在確認・注入手段が未整備
  - Worker Executor は `codex exec ...` を呼び出すが、セッション有無の検証・警告やコンテナへのセッション注入方法（環境変数/ボリューム）が明確でない。
  - 対応: コンテナ起動時にセッション確認を行い、失敗時は UI に警告を返す。セッションの受け渡し仕様をドキュメント化。
