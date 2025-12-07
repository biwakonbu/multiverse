import type { Meta, StoryObj } from '@storybook/svelte';
import WBSGraphNodePreview from './WBSGraphNodePreview.svelte';

const meta = {
  title: 'WBS/WBSGraphNode',
  component: WBSGraphNodePreview,
  tags: ['autodocs'],
  argTypes: {
    type: {
      control: { type: 'select' },
      options: ['phase', 'task'],
    },
    label: { control: 'text' },
    phaseName: {
      control: { type: 'select' },
      options: ['概念設計', '実装設計', '実装', '検証'],
    },
    status: {
      control: { type: 'select' },
      options: ['PENDING', 'READY', 'RUNNING', 'SUCCEEDED', 'COMPLETED', 'FAILED', 'CANCELED', 'BLOCKED'],
    },
    hasChildren: { control: 'boolean' },
    expanded: { control: 'boolean' },
  },
  parameters: {
    layout: 'centered',
  },
} as Meta<typeof WBSGraphNodePreview>;

export default meta;
type Story = StoryObj<typeof meta>;

export const TaskNode: Story = {
  args: {
    type: 'task',
    label: 'Task Node',
    phaseName: '実装',
    status: 'PENDING',
    hasChildren: false,
  },
};

export const TaskRunning: Story = {
  args: {
    type: 'task',
    label: 'Running Task',
    phaseName: '実装',
    status: 'RUNNING',
    hasChildren: false,
  },
};

export const TaskSucceeded: Story = {
  args: {
    type: 'task',
    label: 'Success Task',
    phaseName: '検証',
    status: 'SUCCEEDED',
    hasChildren: false,
  },
};

export const TaskCompleted: Story = {
  args: {
    type: 'task',
    label: 'Completed Task',
    phaseName: '検証',
    status: 'COMPLETED',
    hasChildren: false,
  },
};

export const PhaseNode: Story = {
  args: {
    type: 'phase',
    label: 'Phase Node',
    phaseName: '概念設計',
    status: 'PENDING', // Ignored for phase
    hasChildren: true,
    expanded: true,
  },
};
