# Frontend IDE - Svelte + TypeScript フロントエンド

このディレクトリは multiverse IDE の Web フロントエンドを提供します。

## 責務

- **UI レンダリング**: Svelte コンポーネントによる UI
- **バックエンド通信**: Wails IPC 経由で Go バックエンドを呼び出し
- **状態管理**: Svelte リアクティビティによるローカル状態管理

## ディレクトリ構成

```
frontend/ide/
├── src/
│   ├── main.ts              # エントリポイント
│   ├── app.css              # グローバルスタイル
│   ├── App.svelte           # ルートコンポーネント
│   └── lib/                 # UI コンポーネント
│       ├── WorkspaceSelector.svelte
│       ├── TaskList.svelte
│       ├── TaskDetail.svelte
│       └── TaskCreate.svelte
├── wailsjs/                 # Wails 自動生成バインディング
│   └── go/main/App.js       # Go API の TypeScript ラッパー
├── index.html               # HTML テンプレート
├── package.json             # 依存パッケージ
├── vite.config.ts           # Vite ビルド設定
└── tsconfig.json            # TypeScript 設定
```

## コンポーネント階層

```
App.svelte
├── WorkspaceSelector.svelte  # Workspace 未選択時
└── (Workspace 選択後)
    ├── TaskList.svelte       # サイドバー：タスク一覧
    ├── TaskCreate.svelte     # サイドバー：タスク作成
    └── TaskDetail.svelte     # メイン：タスク詳細
```

## コンポーネント詳細

### App.svelte

ルートコンポーネント。Workspace の選択状態に応じて表示を切り替えます。

**状態**:
- `workspaceId`: 選択中の Workspace ID
- `tasks`: タスク一覧
- `selectedTask`: 選択中のタスク

**ポーリング**:
- 2 秒間隔で `ListTasks()` を呼び出し、タスク状態を更新

### WorkspaceSelector.svelte

Workspace 未選択時に表示。ディレクトリ選択ダイアログを起動します。

**イベント**:
- `selected`: Workspace 選択完了時（詳細: Workspace ID）

### TaskList.svelte

タスク一覧を表示。ステータスに応じた色分け表示。

**Props**:
- `tasks`: タスク配列

**イベント**:
- `select`: タスク選択時（詳細: Task オブジェクト）

**ステータス色**:
- RUNNING: 緑 (#4caf50)
- FAILED: 赤 (#f44336)
- PENDING: オレンジ (#ff9800)

### TaskDetail.svelte

選択中のタスクの詳細を表示。

**Props**:
- `task`: 表示するタスク（null 可）

### TaskCreate.svelte

新規タスク作成フォーム。

**イベント**:
- `created`: タスク作成完了時

## Wails バインディング

### 自動生成ファイル

```
wailsjs/go/main/App.js       # JavaScript バインディング
wailsjs/go/main/App.d.ts     # TypeScript 型定義
```

### 使用例

```typescript
import { ListTasks, CreateTask, SelectWorkspace } from '../wailsjs/go/main/App';

// Workspace 選択
const workspaceId = await SelectWorkspace();

// タスク一覧取得
const tasks = await ListTasks();

// タスク作成
const task = await CreateTask("New Task", "codegen");
```

## 開発コマンド

```bash
# 依存パッケージインストール
npm install

# 開発サーバー起動（Wails dev と連携）
npm run dev

# 本番ビルド
npm run build

# 型チェック
npm run check
```

## 技術スタック

- **Svelte 4**: リアクティブ UI フレームワーク
- **TypeScript 5**: 型安全な JavaScript
- **Vite 5**: 高速ビルドツール
- **Wails v2**: Go ↔ Web IPC

## スタイリング

- コンポーネントスコープ CSS（`<style>` タグ）
- ダークテーマベース（背景 #1e1e1e）
- CSS 変数は未使用（将来的にテーマ対応可能）

## 設計原則

### シンプルな状態管理

- Svelte のリアクティビティを活用
- グローバルストアは使用せず、props/events で伝播
- 複雑になれば `svelte/store` 導入を検討

### ポーリングによる状態同期

- WebSocket 等のリアルタイム通信は未実装
- 2 秒間隔のポーリングで実用的な更新頻度を確保
- 将来的に Server-Sent Events 等への移行を検討

## テスト戦略

### ユニットテスト

- コンポーネントテストは未実装
- 必要に応じて `@testing-library/svelte` 導入

### E2E テスト

- Playwright 等による自動化を検討
- 現状は手動テスト

## 拡張予定

### 短期

- [ ] タスク削除ボタン
- [ ] タスク実行ボタン（TaskDetail 内）
- [ ] ローディング表示

### 長期

- [ ] ログビューア
- [ ] Worker Pool 設定 UI
- [ ] テーマ切り替え（ダーク/ライト）

## 関連ドキュメント

- [../../cmd/multiverse-ide/CLAUDE.md](../../cmd/multiverse-ide/CLAUDE.md): Go バックエンド
- [../../internal/orchestrator/CLAUDE.md](../../internal/orchestrator/CLAUDE.md): Task データモデル
- [../../wails.json](../../wails.json): Wails プロジェクト設定
