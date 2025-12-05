import type { Meta, StoryObj } from '@storybook/svelte';
import Badge from './Badge.svelte';

const meta = {
  title: 'Design System/Badge',
  component: Badge,
  tags: ['autodocs'],
  argTypes: {
    status: {
      control: { type: 'select' },
      options: ['pending', 'ready', 'running', 'succeeded', 'failed', 'canceled', 'blocked'],
      description: 'タスクのステータス'
    },
    size: {
      control: { type: 'select' },
      options: ['small', 'medium'],
      description: 'バッジのサイズ'
    },
    label: {
      control: 'text',
      description: 'カスタムラベル（省略時はステータス名）'
    }
  }
} satisfies Meta<Badge>;

export default meta;
type Story = StoryObj<typeof meta>;

// 各ステータス
export const Pending: Story = {
  args: {
    status: 'pending'
  }
};

export const Ready: Story = {
  args: {
    status: 'ready'
  }
};

export const Running: Story = {
  args: {
    status: 'running'
  }
};

export const Succeeded: Story = {
  args: {
    status: 'succeeded'
  }
};

export const Failed: Story = {
  args: {
    status: 'failed'
  }
};

export const Canceled: Story = {
  args: {
    status: 'canceled'
  }
};

export const Blocked: Story = {
  args: {
    status: 'blocked'
  }
};

// サイズバリエーション
export const Small: Story = {
  args: {
    status: 'running',
    size: 'small'
  }
};

export const Medium: Story = {
  args: {
    status: 'running',
    size: 'medium'
  }
};

// カスタムラベル
export const CustomLabel: Story = {
  args: {
    status: 'running',
    label: '処理中...'
  }
};
