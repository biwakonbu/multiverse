import type { Meta, StoryObj } from "@storybook/svelte";
import TaskBar from "./TaskBar.svelte";

const meta = {
  title: "HUD/TaskBar",
  component: TaskBar,
  tags: ["autodocs"],
  parameters: {
    layout: "fullscreen",
    backgrounds: {
      default: "dark",
      values: [{ name: "dark", value: "#1a1a2e" }],
    },
    docs: {
      description: {
        component:
          "macOS Dock風のタスクバー。各ウィンドウ（Chat, Process, Backlog）のトグルボタンを表示し、ウィンドウの開閉を管理します。",
      },
    },
  },
} as Meta<typeof TaskBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  parameters: {
    docs: {
      description: {
        story:
          "デフォルト状態のタスクバー。画面下部中央に配置され、各ウィンドウのトグルボタンが並びます。",
      },
    },
  },
};
