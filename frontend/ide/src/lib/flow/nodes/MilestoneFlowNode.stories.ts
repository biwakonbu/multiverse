import type { Meta, StoryObj } from '@storybook/svelte';
import MilestoneFlowNode from './MilestoneFlowNode.svelte';

const meta = {
  title: 'Flow/Nodes/MilestoneFlowNode',
  component: MilestoneFlowNode,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
} satisfies Meta<MilestoneFlowNode>;

export default meta;
type Story = StoryObj<typeof meta>;

// Mock WBS Node Data
const createNode = (label: string, percentage: number) => ({
  id: 'm1',
  type: 'milestone',
  label,
  progress: { percentage },
  children: [],
  data: {},
});

export const InProgress: Story = {
  args: {
    data: {
      node: createNode('Alpha Release', 45) as any,
    },
    id: '1',
    position: { x: 0, y: 0 },
    type: 'milestone',
  },
};

export const Completed: Story = {
  args: {
    data: {
      node: createNode('Planning Phase Complete', 100) as any,
    },
    id: '2',
    position: { x: 0, y: 0 },
    type: 'milestone',
  },
};

export const JustStarted: Story = {
  args: {
    data: {
      node: createNode('Project Kickoff', 5) as any,
    },
    id: '3',
    position: { x: 0, y: 0 },
    type: 'milestone',
  },
};
