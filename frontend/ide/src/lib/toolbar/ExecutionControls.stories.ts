import type { Meta, StoryObj } from "@storybook/svelte";
import { fn } from "@storybook/test";
import ExecutionControlsPreview from "./ExecutionControlsPreview.svelte";

const meta = {
  title: "IDE/Toolbar/ExecutionControls",
  component: ExecutionControlsPreview,
  parameters: {
    layout: "centered",
  },
  tags: ["autodocs"],
  argTypes: {
    executionState: {
      control: { type: "radio" },
      options: ["IDLE", "RUNNING", "PAUSED"],
      description: "現在の実行状態",
    },
  },
  args: {
    onStart: fn(),
    onPause: fn(),
    onResume: fn(),
    onStop: fn(),
  },
} as Meta<typeof ExecutionControlsPreview>;

export default meta;
type Story = StoryObj<typeof meta>;

/**
 * 待機状態 - 実行前の状態
 * Startボタンのみが表示される
 */
export const Idle: Story = {
  args: {
    executionState: "IDLE",
  },
};

/**
 * 実行中状態
 * Pause/Stopボタンが表示される
 */
export const Running: Story = {
  args: {
    executionState: "RUNNING",
  },
};

/**
 * 一時停止状態
 * Resume/Stopボタンが表示される
 */
export const Paused: Story = {
  args: {
    executionState: "PAUSED",
  },
};

/**
 * 全ステータス一覧 - 各状態を並べて比較
 */
export const AllStates: Story = {
  render: () => ({
    Component: ExecutionControlsPreview,
    props: { executionState: "IDLE" },
  }),
  decorators: [
    () => ({
      Component: ExecutionControlsPreview,
      props: {},
      template: `
        <div style="display: flex; flex-direction: column; gap: 24px; padding: 24px; background: var(--mv-color-bg-primary);">
          <div style="display: flex; align-items: center; gap: 16px;">
            <span style="width: 80px; color: var(--mv-color-text-secondary); font-size: 12px;">IDLE:</span>
            <slot />
          </div>
        </div>
      `,
    }),
  ],
  parameters: {
    docs: {
      description: {
        story: "全ての実行状態を一覧で確認できます。",
      },
    },
  },
};
