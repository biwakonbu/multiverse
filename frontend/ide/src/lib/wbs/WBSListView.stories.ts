import type { Meta, StoryObj } from '@storybook/svelte';
import WBSListViewPreview from './WBSViewPreview.svelte';

const meta = {
  title: 'WBS/WBSListView',
  component: WBSListViewPreview,
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
          'WBS リストビュー。タスクをフェーズ別にグループ化し、進捗率を表示します。カスタムスクロールバー付き。',
      },
    },
  },
} as Meta<typeof WBSListViewPreview>;

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

// 多数タスク
export const ManyTasks: Story = {
  args: {
    taskCount: 15,
    completedRatio: 0.6,
    showAllPhases: true,
  },
  parameters: {
    docs: {
      description: {
        story: 'スクロールが必要な多数のタスク表示',
      },
    },
  },
};

// 全完了
export const AllCompleted: Story = {
  args: {
    taskCount: 8,
    completedRatio: 1,
    showAllPhases: true,
  },
  parameters: {
    docs: {
      description: {
        story: '全タスク完了状態',
      },
    },
  },
};
