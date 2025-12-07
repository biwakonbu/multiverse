import type { Meta, StoryObj } from "@storybook/svelte";
import ChatInput from "./ChatInput.svelte";

const meta = {
  title: "Chat/ChatInput",
  component: ChatInput,
  tags: ["autodocs"],
  argTypes: {
    disabled: {
      control: "boolean",
      description: "入力を無効化",
    },
  },
  parameters: {
    layout: "centered",
    backgrounds: {
      default: "dark",
      values: [{ name: "dark", value: "#2E3440" }],
    },
    docs: {
      description: {
        component:
          "チャット入力フォーム。ターミナル風のプロンプトアイコン付き。Enterキーで送信、Shift+Enterで改行。",
      },
    },
  },
} as Meta<typeof ChatInput>;

export default meta;
type Story = StoryObj<typeof meta>;

// デフォルト状態
export const Default: Story = {
  args: {
    disabled: false,
  },
};

// 無効状態
export const Disabled: Story = {
  args: {
    disabled: true,
  },
  parameters: {
    docs: {
      description: {
        story:
          "入力が無効化された状態。メッセージ送信中や処理中に使用します。",
      },
    },
  },
};

// インタラクション用（入力テスト）
export const Interactive: Story = {
  args: {
    disabled: false,
  },
  parameters: {
    docs: {
      description: {
        story:
          "インタラクティブなテスト用。テキストを入力してEnterキーで送信イベントが発火します。",
      },
    },
  },
};

// フォーカス状態のデモ
export const Focused: Story = {
  args: {
    disabled: false,
  },
  parameters: {
    docs: {
      description: {
        story: "フォーカス状態。入力フィールドをクリックしてテストしてください。",
      },
    },
  },
  play: async ({ canvasElement }: { canvasElement: HTMLElement }) => {
    // Storybook play function でフォーカス
    const textarea = canvasElement.querySelector("textarea");
    if (textarea) {
      textarea.focus();
    }
  },
};
