import type { Meta, StoryObj } from '@storybook/svelte';
import type { ComponentProps } from 'svelte';
import TaskNodePreview from './TaskNodePreview.svelte';

const meta = {
  title: 'Flow/Nodes/TaskNode',
  component: TaskNodePreview,
  tags: ['autodocs'],
  argTypes: {
    status: {
      control: { type: 'select' },
      options: ['PENDING', 'READY', 'RUNNING', 'SUCCEEDED', 'COMPLETED', 'FAILED', 'CANCELED', 'BLOCKED'],
      description: 'タスクのステータス'
    },
    phaseName: {
      control: { type: 'select' },
      options: ['概念設計', '実装設計', '実装', '検証'],
      description: 'フェーズ名'
    },
    title: {
      control: 'text',
      description: 'タスクのタイトル'
    },
    poolId: {
      control: 'text',
      description: 'ワーカープールID'
    },
    zoomLevel: {
      control: { type: 'range', min: 0.25, max: 3, step: 0.1 },
      description: 'ズームレベル（0.4未満でタイトル非表示、1.2以上で詳細表示）'
    },
    selected: {
      control: 'boolean',
      description: '選択状態'
    },
    dependencies: {
      control: 'object',
      description: '依存タスクID配列'
    }
  },
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: 'Svelte Flow用のTaskNode。GridNodeと同一のビジュアルスタイルを継承。'
      }
    }
  }
} satisfies Meta<typeof TaskNodePreview>;

export default meta;
type Story = StoryObj<typeof meta>;

type TaskNodeArgs = ComponentProps<typeof TaskNodePreview>;

// 各ステータス
export const Pending: Story = {
  args: {
    id: 'task-pending',
    title: 'API エンドポイント設計',
    status: 'PENDING',
    phaseName: '概念設計',
    poolId: 'codegen',
    zoomLevel: 1.5,
    selected: false,
    dependencies: []
  },
  decorators: [
    (_: unknown, { args }: { args: TaskNodeArgs }) => ({
      Component: TaskNodePreview,
      props: args
    })
  ]
};

export const Ready: Story = {
  args: {
    ...Pending.args,
    id: 'task-ready',
    title: 'ユーザー認証機能実装',
    status: 'READY',
    phaseName: '実装設計'
  },
  decorators: Pending.decorators
};

export const Running: Story = {
  args: {
    ...Pending.args,
    id: 'task-running',
    title: 'データベーススキーマ設計',
    status: 'RUNNING',
    phaseName: '実装'
  },
  decorators: Pending.decorators,
  parameters: {
    docs: {
      description: {
        story: '実行中のノードはパルスアニメーションで視覚的にアクティブ状態を表現します。'
      }
    }
  }
};

export const Succeeded: Story = {
  args: {
    ...Pending.args,
    id: 'task-succeeded',
    title: 'テストスイート作成',
    status: 'SUCCEEDED',
    phaseName: '検証'
  },
  decorators: Pending.decorators
};

export const Failed: Story = {
  args: {
    ...Pending.args,
    id: 'task-failed',
    title: 'CI/CD パイプライン構築',
    status: 'FAILED'
  },
  decorators: Pending.decorators
};

export const Blocked: Story = {
  args: {
    ...Pending.args,
    id: 'task-blocked',
    title: '外部API連携',
    status: 'BLOCKED'
  },
  decorators: Pending.decorators
};

// 選択状態
export const Selected: Story = {
  args: {
    ...Pending.args,
    id: 'task-selected',
    title: '選択中のタスク',
    status: 'RUNNING',
    selected: true
  },
  decorators: Pending.decorators,
  parameters: {
    docs: {
      description: {
        story: '選択状態のノードはグローとアウトラインで強調表示されます。'
      }
    }
  }
};

// ズームレベルによる表示変化
export const ZoomLow: Story = {
  args: {
    ...Pending.args,
    title: '低ズーム：タイトル非表示',
    status: 'RUNNING',
    zoomLevel: 0.3
  },
  decorators: Pending.decorators,
  parameters: {
    docs: {
      description: {
        story: 'ズームレベル0.4未満ではタイトルが非表示になり、ステータスのみ表示されます。'
      }
    }
  }
};

export const ZoomMedium: Story = {
  args: {
    ...Pending.args,
    title: '中ズーム：タイトル表示',
    status: 'RUNNING',
    zoomLevel: 1.0
  },
  decorators: Pending.decorators
};

export const ZoomHigh: Story = {
  args: {
    ...Pending.args,
    title: '高ズーム：詳細表示',
    status: 'RUNNING',
    zoomLevel: 1.5
  },
  decorators: Pending.decorators,
  parameters: {
    docs: {
      description: {
        story: 'ズームレベル1.2以上ではプールIDなどの詳細情報も表示されます。'
      }
    }
  }
};

// フェーズ別
export const PhaseConcept: Story = {
  args: {
    ...Pending.args,
    title: '概念設計フェーズ',
    phaseName: '概念設計'
  },
  decorators: Pending.decorators
};

export const PhaseDesign: Story = {
  args: {
    ...Pending.args,
    title: '実装設計フェーズ',
    phaseName: '実装設計'
  },
  decorators: Pending.decorators
};

export const PhaseImpl: Story = {
  args: {
    ...Pending.args,
    title: '実装フェーズ',
    phaseName: '実装'
  },
  decorators: Pending.decorators
};

export const PhaseVerify: Story = {
  args: {
    ...Pending.args,
    title: '検証フェーズ',
    phaseName: '検証'
  },
  decorators: Pending.decorators
};

// 依存関係あり
export const WithDependencies: Story = {
  args: {
    ...Pending.args,
    title: '依存関係のあるタスク',
    dependencies: ['task-1', 'task-2']
  },
  decorators: Pending.decorators,
  parameters: {
    docs: {
      description: {
        story: '依存タスクがある場合、フッターに依存数が表示されます。'
      }
    }
  }
};

// 長いタイトル
export const LongTitle: Story = {
  args: {
    ...Pending.args,
    title: 'これは非常に長いタスクタイトルで省略表示をテストするためのものです。3行を超える内容は省略記号で切り詰められます。',
    status: 'READY',
    zoomLevel: 1.5
  },
  decorators: Pending.decorators,
  parameters: {
    docs: {
      description: {
        story: '長いタイトルは3行まで表示され、それ以上は省略記号（...）で切り詰められます。'
      }
    }
  }
};
