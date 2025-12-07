import type { Meta, StoryObj } from '@storybook/svelte';
import Badge from './Badge.svelte';

const meta = {
  title: 'Design System/Components/Badge',
  component: Badge,
  tags: ['autodocs'],
  argTypes: {
    status: { 
      control: 'select', 
      options: ['pending', 'ready', 'running', 'succeeded', 'completed', 'failed', 'canceled', 'blocked', 'retry_wait'] 
    },
    size: { 
      control: 'select', 
      options: ['small', 'medium'] 
    },
    label: { control: 'text' },
  },
  args: {
    status: 'pending',
    size: 'medium',
  }
} satisfies Meta<Badge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Running: Story = {
  args: {
    status: 'running',
  }
};

export const Succeeded: Story = {
  args: {
    status: 'succeeded',
  }
};

export const Completed: Story = {
  args: {
    status: 'completed',
  }
};

export const Failed: Story = {
  args: {
    status: 'failed',
  }
};

export const RetryWait: Story = {
  args: {
    status: 'retry_wait',
  }
};

export const Small: Story = {
  args: {
    status: 'ready',
    size: 'small',
  }
};
