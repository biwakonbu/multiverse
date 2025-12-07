import type { Meta, StoryObj } from '@storybook/svelte-vite';
import ConnectionLinePreview from './ConnectionLinePreview.svelte';

const meta = {
  title: 'Grid/ConnectionLine',
  component: ConnectionLinePreview,
  tags: ['autodocs'],
  argTypes: {
    satisfied: {
      control: 'boolean',
      description: '依存が満たされているか',
    },
  },
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component:
          'タスク間の依存関係を視覚化する接続線。依存が満たされている場合は緑の実線、未満の場合は赤の破線（アニメーション付き）で表示します。',
      },
    },
  },
} satisfies Meta<ConnectionLinePreview>;

export default meta;
type Story = StoryObj<typeof meta>;

// 満たされた依存
export const Satisfied: Story = {
  args: {
    satisfied: true,
  },
  parameters: {
    docs: {
      description: {
        story:
          '依存元タスクが完了している場合。緑の実線で接続され、依存先タスクは実行可能な状態です。',
      },
    },
  },
};

// 未満の依存
export const Unsatisfied: Story = {
  args: {
    satisfied: false,
  },
  parameters: {
    docs: {
      description: {
        story:
          '依存元タスクが未完了の場合。赤の破線（流れるアニメーション）で接続され、依存先タスクはブロック状態です。',
      },
    },
  },
};
