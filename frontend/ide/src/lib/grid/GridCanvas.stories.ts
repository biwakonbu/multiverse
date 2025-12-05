import type { Meta, StoryObj } from '@storybook/svelte';
import GridCanvasPreview from './GridCanvasPreview.svelte';

// サンプルノードデータ
const sampleNodes = [
  { id: 'task-1', title: 'API設計', status: 'SUCCEEDED' as const, poolId: 'codegen', col: 0, row: 0 },
  { id: 'task-2', title: 'データベース設計', status: 'SUCCEEDED' as const, poolId: 'codegen', col: 1, row: 0 },
  { id: 'task-3', title: '認証実装', status: 'RUNNING' as const, poolId: 'codegen', col: 2, row: 0 },
  { id: 'task-4', title: 'フロントエンド実装', status: 'READY' as const, poolId: 'frontend', col: 0, row: 1 },
  { id: 'task-5', title: 'テスト作成', status: 'PENDING' as const, poolId: 'testing', col: 1, row: 1 },
  { id: 'task-6', title: 'CI/CD構築', status: 'BLOCKED' as const, poolId: 'devops', col: 2, row: 1 },
  { id: 'task-7', title: 'ドキュメント', status: 'PENDING' as const, poolId: 'docs', col: 3, row: 0 },
  { id: 'task-8', title: 'レビュー', status: 'FAILED' as const, poolId: 'review', col: 3, row: 1 },
];

const meta = {
  title: 'Grid/GridCanvas',
  component: GridCanvasPreview,
  tags: ['autodocs'],
  argTypes: {
    zoom: {
      control: { type: 'range', min: 0.25, max: 3, step: 0.1 },
      description: 'ズームレベル'
    },
    panX: {
      control: { type: 'number' },
      description: 'X方向のパン位置'
    },
    panY: {
      control: { type: 'number' },
      description: 'Y方向のパン位置'
    }
  },
  parameters: {
    layout: 'fullscreen',
    docs: {
      description: {
        component: `
Factorio風の2D俯瞰グリッドキャンバス。

## 操作方法
- **スクロール**: パン（移動）
- **Ctrl/Cmd + スクロール**: ズーム
- **Shift + ドラッグ** または **中クリック + ドラッグ**: パン
- **+/-キー**: ズームイン/アウト
- **0キー**: ズームリセット
- **ノードをクリック**: 選択/選択解除

## 特徴
- ズームレベルに応じてノード内の情報表示が変化
- 実行中（RUNNING）のノードはパルスアニメーション
- グリッド背景のドットパターンで位置感覚を補助
`
      }
    }
  }
} satisfies Meta<GridCanvasPreview>;

export default meta;
type Story = StoryObj<typeof meta>;

// デフォルト（複数ノード）
export const Default: Story = {
  args: {
    nodes: sampleNodes,
    zoom: 1,
    panX: 32,
    panY: 32,
    selectedId: null
  }
};

// ズームイン
export const ZoomedIn: Story = {
  args: {
    ...Default.args,
    zoom: 1.5
  },
  parameters: {
    docs: {
      description: {
        story: 'ズームレベル1.5で詳細情報が表示されます。'
      }
    }
  }
};

// ズームアウト
export const ZoomedOut: Story = {
  args: {
    ...Default.args,
    zoom: 0.5
  },
  parameters: {
    docs: {
      description: {
        story: 'ズームレベル0.5で全体を俯瞰できます。タイトルは表示されますが、詳細は非表示。'
      }
    }
  }
};

// 最小ズーム
export const MinZoom: Story = {
  args: {
    ...Default.args,
    zoom: 0.25
  },
  parameters: {
    docs: {
      description: {
        story: '最小ズーム（0.25）ではタイトルが非表示になり、ステータスのみ表示されます。'
      }
    }
  }
};

// 選択状態
export const WithSelection: Story = {
  args: {
    ...Default.args,
    selectedId: 'task-3'
  },
  parameters: {
    docs: {
      description: {
        story: 'ノードをクリックすると選択状態になります。選択中のノードは緑のアウトラインで強調されます。'
      }
    }
  }
};

// 空のキャンバス
export const Empty: Story = {
  args: {
    nodes: [],
    zoom: 1,
    panX: 32,
    panY: 32,
    selectedId: null
  },
  parameters: {
    docs: {
      description: {
        story: 'ノードがない状態。グリッド背景のみ表示されます。'
      }
    }
  }
};

// 大量のノード
const statuses = ['PENDING', 'READY', 'RUNNING', 'SUCCEEDED', 'FAILED', 'BLOCKED'] as const;
const pools = ['codegen', 'frontend', 'testing', 'devops'];
const manyNodes = Array.from({ length: 30 }, (_, i) => ({
  id: `task-${i}`,
  title: `タスク ${i + 1}`,
  status: statuses[i % 6],
  poolId: pools[i % 4],
  col: i % 6,
  row: Math.floor(i / 6)
}));

export const ManyNodes: Story = {
  args: {
    nodes: manyNodes,
    zoom: 0.6,
    panX: 32,
    panY: 32,
    selectedId: null
  },
  parameters: {
    docs: {
      description: {
        story: '30個のノードを配置した例。ズームアウトして全体を俯瞰できます。'
      }
    }
  }
};

// 異なるステータス分布
const statusShowcase = [
  { id: 'pending-1', title: 'Pending タスク', status: 'PENDING' as const, poolId: 'pool', col: 0, row: 0 },
  { id: 'ready-1', title: 'Ready タスク', status: 'READY' as const, poolId: 'pool', col: 1, row: 0 },
  { id: 'running-1', title: 'Running タスク', status: 'RUNNING' as const, poolId: 'pool', col: 2, row: 0 },
  { id: 'succeeded-1', title: 'Succeeded タスク', status: 'SUCCEEDED' as const, poolId: 'pool', col: 0, row: 1 },
  { id: 'failed-1', title: 'Failed タスク', status: 'FAILED' as const, poolId: 'pool', col: 1, row: 1 },
  { id: 'canceled-1', title: 'Canceled タスク', status: 'CANCELED' as const, poolId: 'pool', col: 2, row: 1 },
  { id: 'blocked-1', title: 'Blocked タスク', status: 'BLOCKED' as const, poolId: 'pool', col: 3, row: 0 },
];

export const StatusShowcase: Story = {
  args: {
    nodes: statusShowcase,
    zoom: 1.2,
    panX: 32,
    panY: 32,
    selectedId: null
  },
  parameters: {
    docs: {
      description: {
        story: '全7種類のステータスを一覧表示。各ステータスの色とRUNNINGのアニメーションを確認できます。'
      }
    }
  }
};
