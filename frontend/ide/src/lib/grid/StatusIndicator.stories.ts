import type { Meta, StoryObj } from '@storybook/svelte';
import StatusIndicator from './StatusIndicator.svelte';

const meta = {
  title: 'Grid/StatusIndicator',
  component: StatusIndicator,
  tags: ['autodocs'],
  argTypes: {
    status: {
      control: { type: 'select' },
      options: ['pending', 'ready', 'running', 'succeeded', 'failed', 'canceled', 'blocked'],
      description: 'タスクのステータス'
    },
    size: {
      control: { type: 'select' },
      options: ['small', 'medium', 'large'],
      description: 'インジケーターのサイズ'
    },
    showLabel: {
      control: 'boolean',
      description: 'ラベルを表示するか'
    }
  }
} satisfies Meta<StatusIndicator>;

export default meta;
type Story = StoryObj<typeof meta>;

// 各ステータス
export const Pending: Story = {
  args: {
    status: 'pending',
    showLabel: true
  }
};

export const Ready: Story = {
  args: {
    status: 'ready',
    showLabel: true
  }
};

export const Running: Story = {
  args: {
    status: 'running',
    showLabel: true
  }
};

export const Succeeded: Story = {
  args: {
    status: 'succeeded',
    showLabel: true
  }
};

export const Failed: Story = {
  args: {
    status: 'failed',
    showLabel: true
  }
};

export const Canceled: Story = {
  args: {
    status: 'canceled',
    showLabel: true
  }
};

export const Blocked: Story = {
  args: {
    status: 'blocked',
    showLabel: true
  }
};

// サイズバリエーション
export const Small: Story = {
  args: {
    status: 'running',
    size: 'small',
    showLabel: true
  }
};

export const Medium: Story = {
  args: {
    status: 'running',
    size: 'medium',
    showLabel: true
  }
};

export const Large: Story = {
  args: {
    status: 'running',
    size: 'large',
    showLabel: true
  }
};

// ドットのみ（ラベルなし）
export const DotOnly: Story = {
  args: {
    status: 'running',
    size: 'medium',
    showLabel: false
  }
};

// 全ステータス一覧（ドキュメント用）
export const AllStatuses: Story = {
  render: () => ({
    Component: StatusIndicator,
    props: { status: 'pending', showLabel: true }
  }),
  decorators: [
    () => ({
      Component: StatusIndicator,
      props: {}
    })
  ],
  parameters: {
    docs: {
      description: {
        story: '7つのステータスすべてを表示。running のみパルスアニメーションあり。'
      }
    }
  }
};
