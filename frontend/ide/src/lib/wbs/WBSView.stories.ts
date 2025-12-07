import type { Meta, StoryObj } from '@storybook/svelte-vite';
import WBSViewPreview from './WBSViewPreview.svelte';

const meta = {
  title: 'WBS/WBSView',
  component: WBSViewPreview,
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
          'WBS（作業分解構造）ビュー。タスクをフェーズ別にグループ化し、進捗率を表示します。展開/折りたたみ機能で詳細を制御できます。',
      },
    },
  },
} satisfies Meta<WBSViewPreview>;

export default meta;
type Story = StoryObj<typeof meta>;

// デフォルト表示
export const Default: Story = {
  args: {
    taskCount: 8,
    completedRatio: 0.4,
    showAllPhases: true,
  },
  parameters: {
    docs: {
      description: {
        story: '標準的なWBSビュー。複数のフェーズにタスクが分類されています。',
      },
    },
  },
};

// 空の状態
export const Empty: Story = {
  args: {
    taskCount: 0,
    completedRatio: 0,
    showAllPhases: true,
  },
  parameters: {
    docs: {
      description: {
        story: 'タスクがない状態。チャットからタスクを生成するようユーザーに案内します。',
      },
    },
  },
};

// 全て完了
export const AllCompleted: Story = {
  args: {
    taskCount: 10,
    completedRatio: 1.0,
    showAllPhases: true,
  },
  parameters: {
    docs: {
      description: {
        story: '全タスクが完了した状態。進捗率100%。',
      },
    },
  },
};

// 進行中（半分完了）
export const HalfCompleted: Story = {
  args: {
    taskCount: 12,
    completedRatio: 0.5,
    showAllPhases: true,
  },
  parameters: {
    docs: {
      description: {
        story: '約半分のタスクが完了した状態。',
      },
    },
  },
};

// 開始直後
export const JustStarted: Story = {
  args: {
    taskCount: 8,
    completedRatio: 0.1,
    showAllPhases: true,
  },
  parameters: {
    docs: {
      description: {
        story: 'プロジェクト開始直後の状態。ほとんどのタスクが未完了。',
      },
    },
  },
};

// 単一フェーズ
export const SinglePhase: Story = {
  args: {
    taskCount: 5,
    completedRatio: 0.4,
    showAllPhases: false,
  },
  parameters: {
    docs: {
      description: {
        story: '単一フェーズ（実装）のみのシンプルなビュー。',
      },
    },
  },
};

// 大量タスク
export const ManyTasks: Story = {
  args: {
    taskCount: 20,
    completedRatio: 0.3,
    showAllPhases: true,
  },
  parameters: {
    docs: {
      description: {
        story: '多数のタスクがある場合。スクロールが必要になります。',
      },
    },
  },
};
