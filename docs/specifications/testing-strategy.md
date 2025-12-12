# テスト戦略と方針 (Testing Strategy)

## 概要

Multiverse プロジェクトでは、システム全体の信頼性を確保し、開発効率を向上させるために、**包括的な自動テスト環境**を構築しています。
特に、以下の 3 つの層でテストを実施することで、バックエンドのロジック、フロントエンドの UI 動作、そして視覚的なリグレッションを独立して検証します。

## アーキテクチャ

テストアーキテクチャは以下の 3 層で構成されます。

1.  **Backend Integration E2E**: IDE バックエンドからオーケストレーター、エージェント実行までのフローを検証。
2.  **Frontend UI E2E**: Wails フロントエンドの UI ロジックとユーザー操作を検証。
3.  **Visual Regression Testing (VRT)**: コンポーネント単位およびページ単位での視覚的な変化を自動検知。

| 層           | 範囲                                           | 技術スタック               | 目的                                          |
| ------------ | ---------------------------------------------- | -------------------------- | --------------------------------------------- |
| **Backend**  | `ide` (Go) -> `orchestrator` -> `agent-runner` | Go Test, Shell Script Mock | プロセス連携、タスクキュー、状態遷移の検証    |
| **Frontend** | `frontend/ide` (Svelte)                        | Playwright, Wails JS Mock  | UI 描画、イベントハンドリング、画面遷移の検証 |
| **Visual**   | `frontend/ide` Components                      | Storybook, Playwright      | デザイン崩れの検知、UI カタログ管理           |

---

## 1. Backend Integration E2E

### 配置場所

`test/e2e/orchestrator_flow_test.go`

### 設計方針

実際の IDE アプリケーション (`app.go`) と同様のコンポーネント構成（WorkspaceStore, Scheduler, Executor）をテスト内で再現し、**外部プロセスとの連携**を含めた統合テストを行います。

- **モック化**: 実際の `agent-runner` は実行に時間がかかるため、標準入力を消費して即座に成功を返す `mock_runner.sh` を使用します。
- **検証範囲**:
  - タスクの作成とスケジューリング
  - オーケストレータープロセスによるジョブのピックアップ
  - タスクステータスの遷移 (PENDING -> RUNNING -> SUCCEEDED)
  - 成果物ファイルの生成確認

### 実行方法

```bash
go test -v ./test/e2e/...
```

### 1-2. Backend V2 (Chat to Task)

**配置場所**: `internal/chat/handler_test.go` (Unit/Integration)

v2.0 のチャット駆動タスク生成フローは、LLM (Meta-agent) の出力に依存するため、安定した E2E テストが困難です。
したがって、以下の戦略を採用します。

- **モックベース統合テスト**: `ChatHandler` に対し、モック化された Meta-agent から固定の `DecomposeResponse` を返し、適切に `Task` が生成・保存されるかを検証します。
- **カバレッジ**:
  - `decompose` プロトコルによるタスク生成
  - 依存関係（Dependency）の解決
  - `SuggestedImpl` などの V2 フィールドの保存

---

## 2. Frontend UI E2E

### 配置場所

`frontend/ide/tests/`

### 設計方針

Wails アプリケーションのフロントエンド部分はブラウザ技術で動作しますが、バックエンド（Go）に依存しています。この依存を**モック**することで、バックエンドを起動せずに高速な UI テストを実現します。

- **Playwright**: ブラウザ自動操作ツールとして採用。
- **Wails API Mock**: `frontend/ide/src/mocks/wails.js` に `window.runtime` およびバックエンドメソッド（`CreateTask` 等）のモックを実装。
- **Vite Alias**: E2E テスト実行時のみ、Wails 自動生成ファイルへのパスをモックファイルに向けるように `vite.config.ts` を構成。

### 検証範囲

- タスク一覧の描画
- クリエイト・リード・アップデート・デリート (CRUD) の UI 操作フロー
- コンポーネントの状態変化（ローディング、エラー表示等）

### 実行方法

```bash
cd frontend/ide
npm run test:e2e
```

---

## 3. Frontend Visual Testing

### 配置場所

`frontend/ide/src/**/*.stories.ts` (Storybook)
`frontend/ide/tests/vrt` (Playwright VRT)

### 設計方針

UI の変更による意図しないデザイン崩れ（リグレッション）を防ぐため、スナップショット比較を行います。

1.  **Storybook**:

    - 全 UI コンポーネントのカタログ化 (`npm run storybook`)。
    - 各コンポーネントの "States" (Normal, Error, Loading 等) を Story として定義。

2.  **Visual Regression Testing (VRT)**:
    - Playwright を使用して Storybook の各 Story、または実際のページのスナップショットを撮影。
    - 前回のマスター画像（Golden Image）との差分をピクセル単位で比較。

### 実行方法

```bash
cd frontend/ide

# Storybook 起動
npm run storybook

# VRT 実行 (Playwright)
npm run test:vrt
```

## 今後の展望

- **CI 連携**: GitHub Actions 上でこれらのテストをプルリクエストごとに実行する。
- **カバレッジ拡大**: 異常系（タスク失敗、ネットワークエラー）のテストケースを追加する。
