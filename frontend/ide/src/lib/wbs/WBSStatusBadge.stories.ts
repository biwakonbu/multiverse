import type { Meta, StoryObj } from '@storybook/svelte';
import WBSStatusBadge from './WBSStatusBadge.svelte';

const meta = {
  title: 'WBS/WBSStatusBadge',
  component: WBSStatusBadge,
  tags: ['autodocs'],
  argTypes: {
    status: {
      control: 'select',
      options: ['PENDING', 'RUNNING', 'COMPLETED', 'FAILED', 'CANCELLED', 'RETRY_WAIT'],
    },
  },
} satisfies Meta<WBSStatusBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Pending: Story = {
  args: {
    status: 'PENDING',
  },
};

export const Running: Story = {
  args: {
    status: 'RUNNING',
  },
};

export const Completed: Story = {
  args: {
    status: 'COMPLETED',
  },
};

export const Failed: Story = {
  args: {
    status: 'FAILED',
  },
};

export const Cancelled: Story = {
  args: {
    status: 'CANCELLED',
  },
};

export const RetryWait: Story = {
  args: {
    status: 'RETRY_WAIT',
  },
};
