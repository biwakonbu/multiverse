import type { Meta, StoryObj } from '@storybook/svelte';
import WBSFlowNode from './WBSFlowNode.svelte';
import { expandedNodes } from '../../../stores/wbsStore';

const meta = {
  title: 'Flow/Nodes/WBSFlowNode',
  component: WBSFlowNode,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
  args: {
    id: 'node-1',
    position: { x: 0, y: 0 },
    type: 'wbs',
  }
} satisfies Meta<WBSFlowNode>;

export default meta;
type Story = StoryObj<typeof meta>;

// Mock Data Generators
const createPhaseNode = (id: string, label: string, phaseName: string) => ({
    id,
    type: 'phase',
    label,
    phaseName, // concept, design, impl, verify
    children: [], // No children for simple display
    data: {},
});

const createTaskNode = (id: string, label: string, status: string, phaseName = 'impl') => ({
    id,
    type: 'task',
    label,
    phaseName,
    task: { status: status },
    children: [],
    data: {},
});

export const PhaseConcept: Story = {
  args: {
    data: {
      node: createPhaseNode('n1', 'Concept Phase', 'concept') as any,
    },
  },
};

export const PhaseDesign: Story = {
    args: {
      data: {
        node: createPhaseNode('n2', 'Design Phase', 'design') as any,
      },
    },
};

export const TaskPending: Story = {
    args: {
      data: {
        node: createTaskNode('t1', 'Setup Repository', 'PENDING') as any,
      },
    },
};

export const TaskRunning: Story = {
    args: {
      data: {
        node: createTaskNode('t2', 'Implement Login', 'RUNNING') as any,
      },
    },
};

export const TaskCompleted: Story = {
    args: {
      data: {
        node: createTaskNode('t3', 'Database Schema', 'COMPLETED') as any,
      },
    },
};

export const WithChildrenCollapsed: Story = {
    args: {
        data: {
            node: {
                ...createPhaseNode('n3', 'Implementation', 'impl'),
                children: ['t1', 't2']
            } as any
        }
    }
}

export const WithChildrenExpanded: Story = {
    args: {
        data: {
            node: {
                ...createPhaseNode('n4', 'Verification', 'verify'),
                children: ['t3', 't4']
            } as any
        }
    },
    play: () => {
        expandedNodes.add('n4');
    }
}
