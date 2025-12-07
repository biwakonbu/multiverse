import type { Meta, StoryObj } from "@storybook/svelte";
import BacklogItem from "./BacklogItem.svelte";

const meta = {
  title: "IDE/Backlog/BacklogItem",
  component: BacklogItem,
  parameters: {
    layout: "centered",
    backgrounds: {
      default: "dark",
    },
  },
  tags: ["autodocs"],
} as Meta<typeof BacklogItem>;

export default meta;
type Story = StoryObj<typeof meta>;

const baseDate = new Date();

/**
 * 失敗アイテム - テスト失敗などのエラー
 */
export const Failure: Story = {
  args: {
    item: {
      id: "bl-1",
      taskId: "task-1",
      type: "FAILURE",
      title: "ユニットテスト失敗",
      description: "UserService.createUser() のテストが失敗しました。バリデーションエラーが発生しています。",
      priority: 5,
      createdAt: new Date(baseDate.getTime() - 1000 * 60 * 30).toISOString(), // 30分前
    },
  },
};

/**
 * 質問アイテム - 確認が必要な項目
 */
export const Question: Story = {
  args: {
    item: {
      id: "bl-2",
      taskId: "task-2",
      type: "QUESTION",
      title: "API エンドポイントの設計確認",
      description: "RESTful API のエンドポイント設計について、ネスト構造にするか、フラットにするか確認が必要です。",
      priority: 3,
      createdAt: new Date(baseDate.getTime() - 1000 * 60 * 60 * 2).toISOString(), // 2時間前
    },
  },
};

/**
 * ブロッカーアイテム - 進行を妨げる問題
 */
export const Blocker: Story = {
  args: {
    item: {
      id: "bl-3",
      taskId: "task-3",
      type: "BLOCKER",
      title: "外部API認証トークン期限切れ",
      description: "外部決済APIの認証トークンが期限切れのため、全ての決済関連機能がブロックされています。",
      priority: 5,
      createdAt: new Date(baseDate.getTime() - 1000 * 60 * 15).toISOString(), // 15分前
    },
  },
};

/**
 * 低優先度アイテム
 */
export const LowPriority: Story = {
  args: {
    item: {
      id: "bl-4",
      taskId: "task-4",
      type: "QUESTION",
      title: "ドキュメント更新の確認",
      description: "README.md の更新内容について確認をお願いします。",
      priority: 1,
      createdAt: new Date(baseDate.getTime() - 1000 * 60 * 60 * 24).toISOString(), // 1日前
    },
  },
};

/**
 * 高優先度アイテム
 */
export const HighPriority: Story = {
  args: {
    item: {
      id: "bl-5",
      taskId: "task-5",
      type: "FAILURE",
      title: "本番環境デプロイ失敗",
      description: "CI/CD パイプラインでの本番環境へのデプロイが失敗しました。即座に対応が必要です。",
      priority: 5,
      createdAt: new Date(baseDate.getTime() - 1000 * 60 * 5).toISOString(), // 5分前
    },
  },
};

/**
 * 長いタイトルと説明
 */
export const LongContent: Story = {
  args: {
    item: {
      id: "bl-6",
      taskId: "task-6",
      type: "BLOCKER",
      title: "データベースマイグレーションスクリプトの実行時にトランザクションデッドロックが発生",
      description: "複数のテーブルに対する同時更新処理で、外部キー制約のあるテーブル間でデッドロックが発生しています。マイグレーション順序の見直しとトランザクション分割が必要です。また、関連するインデックスの再構築も検討してください。",
      priority: 4,
      createdAt: new Date(baseDate.getTime() - 1000 * 60 * 45).toISOString(),
    },
  },
};
