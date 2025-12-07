import type { Meta, StoryObj } from '@storybook/svelte-vite';
import RecentWorkspaceList from './RecentWorkspaceList.svelte';

const meta = {
  title: 'Welcome/RecentWorkspaceList',
  component: RecentWorkspaceList,
  tags: ['autodocs'],
  argTypes: {
    workspaces: {
      description: 'ワークスペース一覧'
    },
    loading: {
      control: 'boolean',
      description: '読み込み中状態'
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
        component: '最近使用したワークスペースの一覧。空状態、ローディング状態にも対応。'
      }
    }
  }
} satisfies Meta<RecentWorkspaceList>;

export default meta;
type Story = StoryObj<typeof meta>;

// VRT用に固定タイムスタンプを使用（動的な値は視覚回帰テストを不安定にする）
const FIXED_DATE = new Date('2024-01-15T10:00:00Z');

const mockWorkspaces = [
  {
    id: 'workspace-1',
    displayName: 'My Project',
    projectRoot: '/Users/demo/projects/my-project',
    lastOpenedAt: new Date(FIXED_DATE.getTime() - 1000 * 60 * 30).toISOString()
  },
  {
    id: 'workspace-2',
    displayName: 'Another Project',
    projectRoot: '/Users/demo/projects/another-project',
    lastOpenedAt: new Date(FIXED_DATE.getTime() - 1000 * 60 * 60 * 24).toISOString()
  },
  {
    id: 'workspace-3',
    displayName: 'Third Project',
    projectRoot: '/Users/demo/projects/third',
    lastOpenedAt: new Date(FIXED_DATE.getTime() - 1000 * 60 * 60 * 24 * 3).toISOString()
  }
];

export const WithWorkspaces: Story = {
  args: {
    workspaces: mockWorkspaces,
    loading: false
  }
};

export const Empty: Story = {
  args: {
    workspaces: [],
    loading: false
  },
  parameters: {
    docs: {
      description: {
        story: 'ワークスペースがない場合の空状態表示。'
      }
    }
  }
};

export const Loading: Story = {
  args: {
    workspaces: [],
    loading: true
  },
  parameters: {
    docs: {
      description: {
        story: '読み込み中のローディング表示。'
      }
    }
  }
};

export const ManyWorkspaces: Story = {
  args: {
    workspaces: [
      ...mockWorkspaces,
      {
        id: 'workspace-4',
        displayName: 'Fourth Project',
        projectRoot: '/Users/demo/projects/fourth',
        lastOpenedAt: new Date(FIXED_DATE.getTime() - 1000 * 60 * 60 * 24 * 5).toISOString()
      },
      {
        id: 'workspace-5',
        displayName: 'Fifth Project',
        projectRoot: '/Users/demo/projects/fifth',
        lastOpenedAt: new Date(FIXED_DATE.getTime() - 1000 * 60 * 60 * 24 * 7).toISOString()
      },
      {
        id: 'workspace-6',
        displayName: 'Sixth Project',
        projectRoot: '/Users/demo/projects/sixth',
        lastOpenedAt: new Date(FIXED_DATE.getTime() - 1000 * 60 * 60 * 24 * 10).toISOString()
      }
    ],
    loading: false
  },
  parameters: {
    docs: {
      description: {
        story: '多数のワークスペースがある場合はスクロール可能なリストになります。'
      }
    }
  }
};
