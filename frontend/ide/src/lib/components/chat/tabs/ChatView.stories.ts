import type { Meta, StoryObj } from "@storybook/svelte-vite";
import ChatViewPreview from "./ChatViewPreview.svelte";

const meta = {
  title: "Chat/ChatView",
  component: ChatViewPreview,
  tags: ["autodocs"],
  parameters: {
    layout: "centered",
    backgrounds: {
      default: "dark",
      values: [{ name: "dark", value: "#2E3440" }],
    },
    docs: {
      description: {
        component:
          "ChatViewコンポーネントのプレビュー。メッセージリストのスクロール動作を確認できます。",
      },
    },
  },
} satisfies Meta<ChatViewPreview>;

export default meta;
type Story = StoryObj<typeof meta>;

// 少ないメッセージ（スクロールなし）
export const FewMessages: Story = {
  args: {
    messageCount: 3,
  },
  parameters: {
    docs: {
      description: {
        story: "少数のメッセージ。スクロールは発生しません。",
      },
    },
  },
};

// 多くのメッセージ（スクロールあり）
export const ManyMessages: Story = {
  args: {
    messageCount: 15,
  },
  parameters: {
    docs: {
      description: {
        story: "多くのメッセージがある場合。スクロールが発生し、最新メッセージが表示されます。",
      },
    },
  },
};

// 非常に多くのメッセージ（スクロールテスト）
export const ScrollTest: Story = {
  args: {
    messageCount: 30,
  },
  parameters: {
    docs: {
      description: {
        story: "スクロール動作のテスト用。多数のメッセージで正しくスクロールできることを確認します。",
      },
    },
  },
};

// ロード中状態
export const Loading: Story = {
  args: {
    messageCount: 5,
    isLoading: true,
  },
  parameters: {
    docs: {
      description: {
        story: "ロード中のインジケーター表示。",
      },
    },
  },
};

// エラー状態
export const WithError: Story = {
  args: {
    messageCount: 3,
    error: "チャットセッションの接続に失敗しました",
  },
  parameters: {
    docs: {
      description: {
        story: "エラーバナーが表示される状態。",
      },
    },
  },
};
