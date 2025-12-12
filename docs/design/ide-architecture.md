# Multiverse IDE Architecture

## 概要

Multiverse IDE は、開発者が AI エージェント（Worker）と協働するためのデスクトップアプリケーションです。
Frontend は **Svelte 5** を採用し、高度なグラフ描画に **Svelte Flow** を使用しています。Backend との通信は **Wails v2** を介して行われます。

## 技術スタック

| レイヤー               | 技術              | バージョン/備考                         |
| ---------------------- | ----------------- | --------------------------------------- |
| **Frontend Framework** | **Svelte 5**      | Runes ($state, $derived) を全面採用     |
| **Graph UI**           | **Svelte Flow**   | `@xyflow/svelte` v0.1+                  |
| **Desktop Runtime**    | **Wails v2**      | Go + WebView2/WebKit                    |
| **Styling**            | **CSS Modules**   | Scoped CSS, Design Tokens               |
| **State Management**   | **Svelte Stores** | `writable`, `derived` (Svelte 5 と共存) |

## アーキテクチャ構成

```mermaid
graph TD
    subgraph Frontend [Svelte 5 Context]
        App[App.svelte]
        Canvas[UnifiedFlowCanvas.svelte]
        Store[TaskStore / FlowStore]
        Panel[WBS / Chat Panels]
    end

    subgraph Bridge [Wails Runtime]
        Events[Events (On/Emit)]
        Binds[Go Methods]
    end

    subgraph Backend [Go Context]
        AppGo[App.go]
        Orch[Orchestrator]
        Service[TaskService]
    end

    App --> Canvas
    Canvas -->|Svelte Flow| Store
    App -->|Events| Bridge
    Store <-->|Events| Bridge
    Binds --> Backend
    Backend -->|Events| Bridge
```

## 1. Frontend Design (Svelte 5)

### コンポーネント設計 (Runes)

Svelte 5 の Runes 構文 (`$state`, `$derived`, `$props`, `$effect`) を標準として使用しています。

```svelte
<script lang="ts">
  // Props
  let { taskList }: { taskList: Task[] } = $props();

  // State
  let nodes = $state([]);

  // Derived
  let completedCount = $derived(taskList.filter(t => t.status === 'SUCCEEDED').length);

  // Side Effects
  $effect(() => {
    console.log('Task list updated:', taskList);
  });
</script>
```

### 状態管理 (Stores)

グローバルな状態管理には、Svelte 4 互換の `writable` ストアを使用しています。これは Svelte 5 コンポーネント内でも `$` プレフィックス (`$taskStore`) で透過的に利用可能です。

- `stores/taskStore.ts`: タスクおよび依存関係の管理
- `stores/wbsStore.ts`: WBS 表示モードの状態管理
- `stores/logStore.ts`: 実行ログのストリーム管理

### グラフ描画 (Svelte Flow)

タスクグラフの描画には `@xyflow/svelte` を使用しています。

- **UnifiedFlowCanvas.svelte**: Svelte Flow のラッパーコンポーネント。タスクデータを受け取り、フローのノード/エッジに変換して描画します。
- **Custom Nodes**: `lib/flow/nodes/` にタスク表示専用のノードコンポーネントを定義しています。
- **Layout**: `dagre` アルゴリズムを使用して、タスクの依存関係に基づいた自動レイアウトを提供します。

## 2. Backend Integration (Wails)

Go 製バックエンドとは以下の 2 つのパターンで通信します。

### Method Call (Frontend -> Backend)

`wailsjs` 自動生成コードを使用します。

```typescript
import { CreateTask } from "../../wailsjs/go/main/App";

async function handleSubmit(prompt: string) {
  await CreateTask(prompt); // Go メソッド呼び出し
}
```

### Events (Backend -> Frontend)

Wails のイベントシステムを使用して、非同期な状態更新を受け取ります。

- `task:created`: 新しいタスクが生成された
- `task:stateChange`: タスクのステータスが変化した（PENDING -> RUNNING -> SUCCEEDED）
- `task:log`: 実行ログ（stdout/stderr）のストリーム

`stores/taskStore.ts` 内でリスナーを初期化し、ストアを更しています。

```typescript
// stores/taskStore.ts
EventsOn("task:stateChange", (event) => {
  updateTaskStatus(event.taskId, event.newStatus);
});
```

## 3. デザインシステム

`frontend/ide/src/design-system` に定義された CSS 変数とトークンを使用します。

- **Theme**: Nord Deep (Dark mode optimized)
- **Glassmorphism**: 半透明なパネルとブラー効果 (`--mv-glass-bg`)
- **Grid**: 黄金比ベースのグリッドシステム

## ディレクトリ構造

```
frontend/ide/src/
├── lib/
│   ├── flow/          # Svelte Flow 関連 (Nodes, Edges, Layout)
│   ├── components/    # 共有 UI コンポーネント (Window, Button, PropertyPanel)
│   └── wbs/           # WBS リストビュー
├── stores/            # Svelte Stores
├── design-system/     # CSS 変数・トークン
└── App.svelte         # ルートコンポーネント
```

### 主要コンポーネント

- **`UnifiedFlowCanvas.svelte`**: メインのグラフキャンバス。
- **`TaskNode.svelte`**: タスクノード。`SuggestedImpl` の有無を示すインジケータ (IP) を持つ。
- **`TaskPropPanel.svelte`**: 選択中のタスク詳細を表示するパネル。`SuggestedImpl` や `Artifacts` を表示。
