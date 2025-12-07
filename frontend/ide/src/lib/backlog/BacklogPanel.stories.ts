import type { Meta, StoryObj } from "@storybook/svelte-vite";
import BacklogPanelPreview from "./BacklogPanelPreview.svelte";
import type { BacklogItem } from "./BacklogPanelPreview.svelte";

const meta = {
  title: "IDE/Backlog/BacklogPanel",
  component: BacklogPanelPreview,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
    docs: {
      description: {
        component: `
バックログパネルコンポーネント。

タスク実行中に発生した問題・質問・ブロッカーを一覧表示し、解決・削除操作を行えます。

## 機能
- **アイテムタイプ別表示**: FAILURE（失敗）、QUESTION（質問）、BLOCKER（ブロッカー）
- **優先度表示**: 1-5段階で表示
- **解決ダイアログ**: アイテムを解決する際にコメントを追加可能
- **削除機能**: 確認後にアイテムを削除

## 配置
サイドバーとして左側に表示されます（幅: 320px）。
        `,
      },
    },
  },
  argTypes: {
    items: {
      description: "バックログアイテムの配列",
      control: "object",
    },
  },
} satisfies Meta<BacklogPanelPreview>;

export default meta;
type Story = StoryObj<typeof meta>;

// サンプルデータ
const sampleItems: BacklogItem[] = [
  {
    id: "bl-1",
    taskId: "task-1",
    type: "FAILURE",
    title: "テスト実行エラー",
    description: "npm test でエラーが発生しました。依存パッケージのバージョン不整合が原因の可能性があります。",
    priority: 5,
    createdAt: new Date(Date.now() - 3600000).toISOString(),
  },
  {
    id: "bl-2",
    taskId: "task-2",
    type: "QUESTION",
    title: "API エンドポイントの仕様確認",
    description: "認証フローで使用するエンドポイントの仕様が不明確です。",
    priority: 3,
    createdAt: new Date(Date.now() - 7200000).toISOString(),
  },
  {
    id: "bl-3",
    taskId: "task-3",
    type: "BLOCKER",
    title: "外部サービスの停止",
    description: "サードパーティAPIが一時的にダウンしています。復旧を待つ必要があります。",
    priority: 4,
    createdAt: new Date(Date.now() - 1800000).toISOString(),
  },
];

// 空の状態
export const Empty: Story = {
  args: {
    items: [],
  },
  parameters: {
    docs: {
      description: {
        story: "バックログが空の状態。チェックマークと「バックログは空です」メッセージが表示されます。",
      },
    },
  },
};

// 複数アイテムがある状態
export const WithItems: Story = {
  args: {
    items: sampleItems,
  },
  parameters: {
    docs: {
      description: {
        story: "複数のバックログアイテムがある状態。各タイプ（FAILURE、QUESTION、BLOCKER）のスタイルを確認できます。",
      },
    },
  },
};

// 失敗のみ
export const FailuresOnly: Story = {
  args: {
    items: [
      {
        id: "bl-f1",
        taskId: "task-1",
        type: "FAILURE",
        title: "ビルドエラー",
        description: "TypeScript コンパイルに失敗しました。",
        priority: 5,
        createdAt: new Date().toISOString(),
      },
      {
        id: "bl-f2",
        taskId: "task-2",
        type: "FAILURE",
        title: "E2E テスト失敗",
        description: "ログインフローのテストが失敗しました。",
        priority: 4,
        createdAt: new Date(Date.now() - 600000).toISOString(),
      },
    ],
  },
  parameters: {
    docs: {
      description: {
        story: "失敗アイテムのみの状態。左側に赤いボーダーが表示されます。",
      },
    },
  },
};

// 多数のアイテム
export const ManyItems: Story = {
  args: {
    items: [
      ...sampleItems,
      {
        id: "bl-4",
        taskId: "task-4",
        type: "FAILURE",
        title: "Lint エラー",
        description: "ESLint ルール違反が検出されました。",
        priority: 2,
        createdAt: new Date(Date.now() - 10800000).toISOString(),
      },
      {
        id: "bl-5",
        taskId: "task-5",
        type: "QUESTION",
        title: "デザインの確認",
        description: "ボタンの色を変更するべきか？",
        priority: 1,
        createdAt: new Date(Date.now() - 14400000).toISOString(),
      },
      {
        id: "bl-6",
        taskId: "task-6",
        type: "BLOCKER",
        title: "権限不足",
        description: "デプロイに必要な権限がありません。",
        priority: 5,
        createdAt: new Date(Date.now() - 900000).toISOString(),
      },
    ],
  },
  parameters: {
    docs: {
      description: {
        story: "多数のアイテムがある状態。スクロールが有効になります。",
      },
    },
  },
};
