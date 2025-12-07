import type { Meta, StoryObj } from "@storybook/svelte";
import StoreWrapper from "./StoreWrapper.svelte";

const meta = {
  title: "Utils/StoreWrapper",
  component: StoreWrapper,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
    docs: {
      description: {
        component:
          "Storybook用のストアラッパー。ストーリー内でwindowStoreの初期状態を設定するために使用します。",
      },
    },
  },
} as Meta<typeof StoreWrapper>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    initialState: {
      chat: {
        isOpen: true,
        position: { x: 50, y: 50 },
        zIndex: 100,
      },
    },
  },
  parameters: {
    docs: {
      description: {
        story: "デフォルト状態のストアラッパー。",
      },
    },
  },
};
