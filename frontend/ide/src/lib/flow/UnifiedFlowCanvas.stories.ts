import type { Meta, StoryObj } from '@storybook/svelte';
import UnifiedFlowCanvas from './UnifiedFlowCanvas.svelte';
import type { Task } from '../../types';

const meta = {
  title: 'Flow/UnifiedFlowCanvas',
  component: UnifiedFlowCanvas,
  tags: ['autodocs'],
  argTypes: {
    taskList: { control: 'object' },
  },
} satisfies Meta<UnifiedFlowCanvas>;

export default meta;
type Story = StoryObj<typeof meta>;

const mockTasks: Task[] = [
  {
    id: '1',
    title: 'Task 1',
    status: 'COMPLETED',
    phaseName: '概念設計',
    dependencies: [],
    poolId: 'p1',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: '2',
    title: 'Task 2',
    status: 'RUNNING',
    phaseName: '実装設計',
    dependencies: ['1'],
    poolId: 'p1',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
  {
    id: '3',
    title: 'Task 3',
    status: 'PENDING',
    phaseName: '実装',
    dependencies: ['2'],
    poolId: 'p1',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  },
];

export const Default: Story = {
  args: {
    taskList: mockTasks,
  },
};

export const Empty: Story = {
  args: {
    taskList: [],
  },
};
