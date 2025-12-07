import type { Meta, StoryObj } from '@storybook/svelte-vite';
import WorkspaceCard from './WorkspaceCard.svelte';

// VRT用に固定タイムスタンプを使用（動的な値は視覚回帰テストを不安定にする）
const FIXED_DATE = new Date('2024-01-15T10:00:00Z');

const meta = {
  title: 'Welcome/WorkspaceCard',
  component: WorkspaceCard,
  tags: ['autodocs'],
  argTypes: {
    workspace: {
      description: 'ワークスペース情報'
    }
  },
  parameters: {
    layout: 'centered',
    backgrounds: {
      default: 'dark',
      values: [
        { name: 'dark', value: '#16181e' }
      ]
    },
    docs: {
      description: {
        component: '個別のワークスペースを表示するカード。クリックで開く、×ボタンで削除。'
      }
    }
  }
} satisfies Meta<WorkspaceCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    workspace: {
      id: 'workspace-1',
      displayName: 'My Project',
      projectRoot: '/Users/demo/projects/my-project',
      lastOpenedAt: FIXED_DATE.toISOString()
    }
  }
};

export const LongPath: Story = {
  args: {
    workspace: {
      id: 'workspace-2',
      displayName: 'Very Long Project Name That Should Be Truncated',
      projectRoot: '/Users/demo/very/long/path/to/a/deeply/nested/project/directory',
      lastOpenedAt: new Date(FIXED_DATE.getTime() - 1000 * 60 * 60 * 24).toISOString()
    }
  },
  parameters: {
    docs: {
      description: {
        story: '長いパスやプロジェクト名は省略記号で切り詰められます。'
      }
    }
  }
};

export const RecentlyUsed: Story = {
  args: {
    workspace: {
      id: 'workspace-3',
      displayName: 'Recent Project',
      projectRoot: '/Users/demo/projects/recent',
      lastOpenedAt: new Date(FIXED_DATE.getTime() - 1000 * 60 * 5).toISOString() // 5分前
    }
  },
  parameters: {
    docs: {
      description: {
        story: '最近使用したワークスペース。'
      }
    }
  }
};

export const OldWorkspace: Story = {
  args: {
    workspace: {
      id: 'workspace-4',
      displayName: 'Old Project',
      projectRoot: '/Users/demo/projects/old',
      lastOpenedAt: new Date(FIXED_DATE.getTime() - 1000 * 60 * 60 * 24 * 30).toISOString() // 30日前
    }
  },
  parameters: {
    docs: {
      description: {
        story: '古いワークスペース。最終使用日が表示されます。'
      }
    }
  }
};
