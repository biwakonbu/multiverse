import type { Meta, StoryObj } from '@storybook/svelte-vite';
import WBSHeaderPreview from './WBSHeaderPreview.svelte';

const meta = {
  title: 'WBS/WBSHeader',
  component: WBSHeaderPreview,
  tags: ['autodocs'],
  argTypes: {
    percentage: {
      control: { type: 'range', min: 0, max: 100, step: 1 },
      description: '進捗率 (%)',
    },
    completed: {
      control: { type: 'number', min: 0 },
      description: '完了タスク数',
    },
    total: {
      control: { type: 'number', min: 0 },
      description: '総タスク数',
    },
  },
  parameters: {
    layout: 'centered',
  },
} satisfies Meta<WBSHeaderPreview>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Initial: Story = {
  args: {
    percentage: 0,
    completed: 0,
    total: 10,
  },
};

export const InProgressLow: Story = {
  args: {
    percentage: 20,
    completed: 2,
    total: 10,
  },
};

export const InProgressMid: Story = {
  args: {
    percentage: 55,
    completed: 11,
    total: 20,
  },
};

export const InProgressHigh: Story = {
  args: {
    percentage: 85,
    completed: 17,
    total: 20,
  },
};

export const Completed: Story = {
  args: {
    percentage: 100,
    completed: 20,
    total: 20,
  },
};
