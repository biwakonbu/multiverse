import type { Meta, StoryObj } from '@storybook/svelte';
import WBSNodePreview from './WBSNodePreview.svelte';

const meta = {
  title: 'WBS/WBSNode',
  component: WBSNodePreview,
  tags: ['autodocs'],
  argTypes: {
    type: {
      control: { type: 'select' },
      options: ['phase', 'task'],
      description: 'ノードタイプ（フェーズまたはタスク）',
    },
    label: {
      control: 'text',
      description: 'ノードのラベル',
    },
    phaseName: {
      control: { type: 'select' },
      options: ['概念設計', '実装設計', '実装', '検証'],
      description: 'フェーズ名',
    },
    status: {
      control: { type: 'select' },
      options: ['PENDING', 'READY', 'RUNNING', 'SUCCEEDED', 'COMPLETED', 'FAILED', 'CANCELED', 'BLOCKED'],
      description: 'タスクのステータス',
    },
    level: {
      control: { type: 'number', min: 0, max: 5 },
      description: 'インデントレベル',
    },
    completed: {
      control: { type: 'number', min: 0 },
      description: '完了タスク数（フェーズのみ）',
    },
    total: {
      control: { type: 'number', min: 0 },
      description: '総タスク数（フェーズのみ）',
    },
    expanded: {
      control: 'boolean',
      description: '展開状態（フェーズのみ）',
    },
    hasChildren: {
      control: 'boolean',
      description: '子ノードがあるか',
    },
  },
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component:
          'WBS（作業分解構造）のノードコンポーネント。フェーズノードとタスクノードの2種類があり、それぞれ異なる表示をします。',
      },
    },
  },
} as Meta<typeof WBSNodePreview>;

export default meta;
type Story = StoryObj<typeof meta>;

// フェーズノード
export const PhaseNode: Story = {
  args: {
    type: 'phase',
    label: '概念設計',
    phaseName: '概念設計',
    level: 0,
    completed: 3,
    total: 5,
    expanded: true,
    hasChildren: true,
  },
  parameters: {
    docs: {
      description: {
        story: 'フェーズノードは進捗バーと展開/折りたたみトグルを表示します。',
      },
    },
  },
};

export const PhaseNodeCollapsed: Story = {
  args: {
    ...PhaseNode.args,
    expanded: false,
  },
  parameters: {
    docs: {
      description: {
        story: '折りたたまれた状態のフェーズノード。',
      },
    },
  },
};

export const PhaseNodeComplete: Story = {
  args: {
    type: 'phase',
    label: '実装設計',
    phaseName: '実装設計',
    level: 0,
    completed: 4,
    total: 4,
    expanded: true,
    hasChildren: true,
  },
  parameters: {
    docs: {
      description: {
        story: '全タスク完了のフェーズノード（進捗100%）。',
      },
    },
  },
};

// タスクノード - 各ステータス
export const TaskPending: Story = {
  args: {
    type: 'task',
    label: 'APIエンドポイント設計',
    phaseName: '概念設計',
    status: 'PENDING',
    level: 1,
  },
};

export const TaskReady: Story = {
  args: {
    type: 'task',
    label: 'データベーススキーマ設計',
    phaseName: '概念設計',
    status: 'READY',
    level: 1,
  },
};

export const TaskRunning: Story = {
  args: {
    type: 'task',
    label: 'ユーザー認証機能実装',
    phaseName: '実装',
    status: 'RUNNING',
    level: 1,
  },
  parameters: {
    docs: {
      description: {
        story: '実行中のタスク。',
      },
    },
  },
};

export const TaskSucceeded: Story = {
  args: {
    type: 'task',
    label: 'ログイン画面作成',
    phaseName: '実装',
    status: 'SUCCEEDED',
    level: 1,
  },
};

export const TaskCompleted: Story = {
  args: {
    type: 'task',
    label: '完了済みタスク',
    phaseName: '検証',
    status: 'COMPLETED',
    level: 1,
  },
};

export const TaskFailed: Story = {
  args: {
    type: 'task',
    label: 'パフォーマンステスト',
    phaseName: '検証',
    status: 'FAILED',
    level: 1,
  },
};

export const TaskCanceled: Story = {
  args: {
    type: 'task',
    label: 'レガシーAPI対応',
    phaseName: '実装',
    status: 'CANCELED',
    level: 1,
  },
};

export const TaskBlocked: Story = {
  args: {
    type: 'task',
    label: '外部サービス連携',
    phaseName: '実装',
    status: 'BLOCKED',
    level: 1,
  },
  parameters: {
    docs: {
      description: {
        story: '依存タスクが未完了のためブロックされているタスク。',
      },
    },
  },
};

// インデントレベル
export const NestedLevel2: Story = {
  args: {
    type: 'task',
    label: 'サブタスク（レベル2）',
    phaseName: '実装',
    status: 'PENDING',
    level: 2,
  },
  parameters: {
    docs: {
      description: {
        story: 'インデントレベル2のタスク。',
      },
    },
  },
};

// 長いラベル
export const LongLabel: Story = {
  args: {
    type: 'task',
    label: 'これは非常に長いタスク名で省略表示をテストするためのものです。文字数が多いとどう表示されるか確認します。',
    phaseName: '実装',
    status: 'RUNNING',
    level: 1,
  },
  parameters: {
    docs: {
      description: {
        story: '長いラベルは省略記号（...）で切り詰められます。',
      },
    },
  },
};
