import type { Meta, StoryObj } from '@storybook/svelte-vite';
import StatusBadge from './StatusBadge.svelte';

const meta = {
  title: 'WBS/StatusBadge',
  component: StatusBadge,
  tags: ['autodocs'],
  argTypes: {
    status: {
      control: { type: 'select' },
      options: ['PENDING', 'READY', 'RUNNING', 'SUCCEEDED', 'COMPLETED', 'FAILED', 'CANCELED', 'BLOCKED'],
      description: 'タスクのステータス',
    },
  },
  parameters: {
    layout: 'centered',
  },
} satisfies Meta<StatusBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Pending: Story = { args: { status: 'PENDING' } };
export const Ready: Story = { args: { status: 'READY' } };
export const Running: Story = { args: { status: 'RUNNING' } };
export const Succeeded: Story = { args: { status: 'SUCCEEDED' } };
export const Completed: Story = { args: { status: 'COMPLETED' } };
export const Failed: Story = { args: { status: 'FAILED' } };
export const Canceled: Story = { args: { status: 'CANCELED' } };
export const Blocked: Story = { args: { status: 'BLOCKED' } };
