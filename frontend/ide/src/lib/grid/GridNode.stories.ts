import type { Meta, StoryObj } from '@storybook/svelte';
import type { ComponentProps } from 'svelte';
import GridNodePreview from './GridNodePreview.svelte';

const meta = {
  title: 'Grid/GridNode',
  component: GridNodePreview,
  tags: ['autodocs'],
  argTypes: {
    status: {
      control: { type: 'select' },
      options: ['PENDING', 'READY', 'RUNNING', 'SUCCEEDED', 'FAILED', 'CANCELED', 'BLOCKED'],
      description: 'タスクのステータス'
    },
    title: {
      control: 'text',
      description: 'タスクのタイトル'
    },
    poolId: {
      control: 'text',
      description: 'ワーカープールID'
    },
    col: {
      control: { type: 'number', min: 0, max: 10 },
      description: 'グリッド列位置'
    },
    row: {
      control: { type: 'number', min: 0, max: 10 },
      description: 'グリッド行位置'
    },
    zoomLevel: {
      control: { type: 'range', min: 0.25, max: 3, step: 0.1 },
      description: 'ズームレベル（0.4未満でタイトル非表示、1.2以上で詳細表示）'
    },
    selected: {
      control: 'boolean',
      description: '選択状態'
    }
  },
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: 'Factorio風グリッドUI上のタスクノード。ステータスに応じた色とアニメーションを表示。'
      }
    }
  },
  decorators: [
    () => ({
      Component: GridNodePreview,
      props: {},
      // ノードの位置を見やすくするためのラッパー
    })
  ]
} as Meta<typeof GridNodePreview>;

export default meta;
type Story = StoryObj<typeof meta>;

type GridNodeArgs = ComponentProps<typeof GridNodePreview>;

// 各ステータス
export const Pending: Story = {
  args: {
    id: 'task-pending',
    title: 'API エンドポイント設計',
    status: 'PENDING',
    poolId: 'codegen',
    col: 0,
    row: 0,
    zoomLevel: 1.5,
    selected: false
  },
  decorators: [
    (_: unknown, { args }: { args: GridNodeArgs }) => ({
      Component: GridNodePreview,
      props: args
    })
  ]
};

export const Ready: Story = {
  args: {
    ...Pending.args,
    id: 'task-ready',
    title: 'ユーザー認証機能実装',
    status: 'READY'
  },
  decorators: Pending.decorators
};

export const Running: Story = {
  args: {
    ...Pending.args,
    id: 'task-running',
    title: 'データベーススキーマ設計',
    status: 'RUNNING'
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
    status: 'SUCCEEDED'
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

export const Canceled: Story = {
  args: {
    ...Pending.args,
    id: 'task-canceled',
    title: 'レガシーコード移行',
    status: 'CANCELED'
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
        story: '選択状態のノードは緑のアウトラインで強調表示されます。'
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

// 長いタイトル
export const LongTitle: Story = {
  args: {
    ...Pending.args,
    title: 'これは非常に長いタスクタイトルで省略表示をテストするためのものです',
    status: 'READY',
    zoomLevel: 1.5
  },
  decorators: Pending.decorators,
  parameters: {
    docs: {
      description: {
        story: '長いタイトルは省略記号（...）で切り詰められます。'
      }
    }
  }
};
