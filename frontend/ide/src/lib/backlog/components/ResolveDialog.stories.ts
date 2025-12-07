import type { Meta, StoryObj } from "@storybook/svelte";
import ResolveDialog from "./ResolveDialog.svelte";

const meta = {
  title: "IDE/Backlog/ResolveDialog",
  component: ResolveDialog,
  parameters: {
    layout: "fullscreen",
  },
  tags: ["autodocs"],
} as Meta<typeof ResolveDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

const baseDate = new Date();

/**
 * デフォルト状態 - 失敗アイテムの解決
 */
export const Default: Story = {
  args: {
    item: {
      id: "bl-1",
      taskId: "task-1",
      type: "FAILURE",
      title: "ユニットテスト失敗",
      description: "UserService.createUser() のテストが失敗しました。",
      priority: 5,
      createdAt: baseDate.toISOString(),
    },
  },
};

/**
 * 質問アイテムの解決
 */
export const ResolveQuestion: Story = {
  args: {
    item: {
      id: "bl-2",
      taskId: "task-2",
      type: "QUESTION",
      title: "API エンドポイントの設計確認",
      description: "RESTful API のエンドポイント設計について確認が必要です。",
      priority: 3,
      createdAt: baseDate.toISOString(),
    },
  },
};

/**
 * ブロッカーアイテムの解決
 */
export const ResolveBlocker: Story = {
  args: {
    item: {
      id: "bl-3",
      taskId: "task-3",
      type: "BLOCKER",
      title: "外部API認証トークン期限切れ",
      description: "外部決済APIの認証トークンが期限切れです。",
      priority: 5,
      createdAt: baseDate.toISOString(),
    },
  },
};

/**
 * 長いタイトルのアイテム
 */
export const LongTitle: Story = {
  args: {
    item: {
      id: "bl-4",
      taskId: "task-4",
      type: "FAILURE",
      title: "データベースマイグレーションスクリプトの実行時にトランザクションデッドロックが発生しています",
      description: "複数テーブルへの同時更新処理でデッドロックが発生しています。",
      priority: 4,
      createdAt: baseDate.toISOString(),
    },
  },
};
