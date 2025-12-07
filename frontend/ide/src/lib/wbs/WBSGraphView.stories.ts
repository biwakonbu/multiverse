import type { Meta, StoryObj } from '@storybook/svelte-vite';
import WBSGraphViewPreview from './WBSGraphViewPreview.svelte';

const meta = {
  title: 'WBS/WBSGraphView',
  component: WBSGraphViewPreview,
  tags: ['autodocs'],
  argTypes: {
    taskCount: {
      control: { type: 'range', min: 0, max: 20, step: 1 },
      description: 'サンプルタスク数',
    },
    completedRatio: {
      control: { type: 'range', min: 0, max: 1, step: 0.1 },
      description: '完了率（0〜1）',
    },
    showAllPhases: {
      control: 'boolean',
      description: '全フェーズを表示するか',
    },
  },
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component:
          'WBS グラフビュー。タスクをツリー構造で視覚的に表示します。ドラッグスクロール対応。',
      },
    },
  },
} satisfies Meta<WBSGraphViewPreview>;

export default meta;
type Story = StoryObj<typeof meta>;

// デフォルト表示
export const Default: Story = {
  args: {
    taskCount: 5,
    completedRatio: 0.4,
    showAllPhases: true,
  },
};

// 空状態
export const Empty: Story = {
  args: {
    taskCount: 0,
    completedRatio: 0,
    showAllPhases: false,
  },
  parameters: {
    docs: {
      description: {
        story: 'タスクが0件の場合の空状態表示',
      },
    },
  },
};

// 大規模ツリー
export const LargeTree: Story = {
  args: {
    taskCount: 15,
    completedRatio: 0.5,
    showAllPhases: true,
  },
  parameters: {
    docs: {
      description: {
        story: '大規模なツリー構造の表示',
      },
    },
  },
};
