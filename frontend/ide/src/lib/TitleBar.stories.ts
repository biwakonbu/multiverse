import type { Meta, StoryObj } from "@storybook/svelte";
import TitleBar from "./TitleBar.svelte";

const meta = {
  title: "IDE/TitleBar",
  component: TitleBar,
  parameters: {
    layout: "fullscreen",
  },
  tags: ["autodocs"],
  decorators: [
    () => ({
      Component: TitleBar,
      template: `
        <div style="height: 200px; background: var(--mv-color-surface-app); position: relative;">
          <slot />
          <div style="padding-top: 40px; padding-left: 16px; color: var(--mv-color-text-secondary); font-size: 12px;">
            タイトルバー領域（透明・ドラッグ可能）
          </div>
        </div>
      `,
    }),
  ],
} as Meta<typeof TitleBar>;

export default meta;
type Story = StoryObj<typeof meta>;

/**
 * デフォルト状態
 * 透明なタイトルバー領域。Wails環境ではウィンドウのドラッグ領域として機能する。
 */
export const Default: Story = {};

/**
 * 背景付きでの表示
 * タイトルバーの高さと位置を確認するためのプレビュー
 */
export const WithVisibleBorder: Story = {
  decorators: [
    () => ({
      Component: TitleBar,
      template: `
        <div style="height: 200px; background: var(--mv-color-surface-app); position: relative;">
          <div style="position: fixed; top: 0; left: 0; right: 0; height: 32px; border-bottom: 1px dashed var(--mv-color-border-default); opacity: 0.5;"></div>
          <slot />
          <div style="padding-top: 40px; padding-left: 16px; color: var(--mv-color-text-secondary); font-size: 12px;">
            タイトルバー領域を破線で表示（高さ: 32px）
          </div>
        </div>
      `,
    }),
  ],
};
