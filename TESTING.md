# テストに関する知識とベストプラクティス

## 1. テストの種類

- **ユニットテスト**: 個々の関数やメソッドを対象に、外部依存をモック化して実行します。
- **プロパティベーステスト (PBT)**: `gopter` などのライブラリを使い、入力の範囲を自動生成して不変条件を検証します。テストケース数は `MinSuccessfulTests` で調整可能です。
- **統合テスト**: 複数コンポーネントを組み合わせ、実装をモック化して検証します。
- **Docker Sandbox テスト**: 実際の Docker コンテナでサンドボックス管理の動作を検証します（`-tags=docker` で実行）。
- **Codex 統合テスト**: 実際の Codex CLI を使用した end-to-end テスト（`-tags=codex` で実行）。

## 2. テストの実装ポイント

1. **依存性の抽象化**
   - `MetaClient`, `WorkerExecutor`, `NoteWriter` などはインターフェース化し、テスト時にモック実装 (`internal/mock`) を注入します。
2. \*\*モックの作成
   - `mock.MetaClient` は `PlanTask` と `NextAction` の戻り値を自由に設定でき、シナリオごとに異なる挙動をシミュレートできます。
   - `mock.WorkerExecutor` は `RunWorker` の結果 (`WorkerRunResult`) を固定して返すだけで、実際の Docker コンテナ起動は不要です。
   - `mock.NoteWriter` はファイル書き込みをスキップし、テストの副作用を防ぎます。
3. **PBT の設定**
   - `parameters.MinSuccessfulTests` を適切に設定し、テスト実行時間とカバレッジのバランスを取ります。デバッグ時は 5〜10、CI では 50〜100 が目安です。
   - 生成するデータは `gen.IntRange` や `gen.AnyString` で制限し、極端なケースが原因でテストがハングしないようにします。
4. **テストの実行**
   - ユニットテスト（依存なし）: `go test ./...`
   - Mock 統合テスト: `go test ./test/integration/...`
   - Docker Sandbox テスト: `go test -tags=docker -timeout=10m ./test/sandbox/...`
   - Codex 統合テスト: `go test -tags=codex -timeout=10m ./test/codex/...`
   - 全テスト: `go test -tags=docker,codex -timeout=15m ./...`
   - 並列実行: `go test -parallel 4 ./...`
   - カバレッジ: `go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out`

## 3. トラブルシューティング

- **テストがハングする**
  - PBT のケース数が多すぎる、または生成器が無限ループに陥っている可能性があります。`MinSuccessfulTests` を減らし、`gen` の範囲を狭めて再実行してください。
  - 依存モックが期待通りに呼び出されていない場合、モック実装の `RunWorkerFunc` がブロックしていないか確認します。
- **Docker Sandbox が起動しない**
  - 統合テストで実際のサンドボックスを使用する場合、Docker がインストールされ、現在のユーザーが `docker` グループに所属しているか確認してください。
  - `sandbox.StartContainer` のエラーメッセージをログに出力し、`docker run` のパラメータが正しいか検証します。
- **Mock が期待と違う**
  - `mock.MetaClient` の `PlanTaskFunc` / `NextActionFunc` がテストケースごとに正しく設定されているか、`prop.ForAll` の引数と一致しているか確認します。

## 4. ベストプラクティス

- テストは **高速** に保ち、CI では **並列実行** (`go test -parallel N`) を活用します。
- 失敗したテストは **ログ出力** を充実させ、`t.Fatalf` や `t.Errorf` で詳細情報を残します。
- 重要なロジックは **PBT** で不変条件を検証し、境界条件は手動テストで補完します。
- 依存性注入により、**実装とテストを分離** し、モックの差し替えを容易にします。

---

このドキュメントは `TESTING.md` としてリポジトリのルートに配置し、開発者がテストの書き方やトラブルシューティングをすぐに参照できるようにしてください。
