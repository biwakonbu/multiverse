
import type { Meta, StoryObj } from '@storybook/svelte';
import ResourceList from './ResourceList.svelte';
import type { ResourceNode } from './types';

const meta = {
  title: 'HUD/ResourceList',
  component: ResourceList,
  tags: ['autodocs'],
  parameters: {
      backgrounds: {
          default: 'dark',
      }
  }
} as Meta<typeof ResourceList>;

export default meta;
type Story = StoryObj<typeof meta>;

const mockData: ResourceNode[] = [
    {
        id: 'sys',
        name: 'Multiverse Orchestrator',
        type: 'ORCHESTRATOR',
        status: 'RUNNING',
        detail: 'Tasks: 1 Active, 2 Queued',
        expanded: true,
        children: [
            {
                id: 'meta-1',
                name: 'Meta-Agent: Plan Refiner',
                type: 'META',
                status: 'THINKING',
                detail: 'Analyzing user request...',
                expanded: true,
                children: [
                    {
                        id: 'worker-1',
                        name: 'Worker: Codex-01',
                        type: 'WORKER',
                        status: 'RUNNING',
                        detail: 'Executing: go test ./...',
                        expanded: true,
                        children: [
                           {
                               id: 'container-1',
                               name: 'Container: vigorous_turing',
                               type: 'CONTAINER',
                               status: 'RUNNING',
                               detail: 'Image: biwakonbu/multiverse-worker:latest'
                           } 
                        ]
                    }
                ]
            }
        ]
    }
];

const mockDataIdle: ResourceNode[] = [
    {
        id: 'sys',
        name: 'Multiverse Orchestrator',
        type: 'ORCHESTRATOR',
        status: 'IDLE',
        detail: 'Waiting for tasks...',
        children: []
    }
];

const mockDataError: ResourceNode[] = [
    {
        id: 'sys',
        name: 'Multiverse Orchestrator',
        type: 'ORCHESTRATOR',
        status: 'RUNNING',
        expanded: true,
        children: [
            {
                id: 'meta-1',
                name: 'Meta-Agent: Bug Fixer',
                type: 'META',
                status: 'ERROR',
                detail: 'Failed to parse prompt',
                expanded: true,
                children: [
                    {
                        id: 'worker-1',
                        name: 'Worker: Codex-02',
                        type: 'WORKER',
                        status: 'IDLE',
                        detail: 'Terminated unexpectedly'
                    }
                ]
            }
        ]
    }
];

export const SystemRunning: Story = {
  args: {
    resources: mockData,
  },
};

export const SystemIdle: Story = {
  args: {
    resources: mockDataIdle,
  },
};

export const SystemError: Story = {
    args: {
        resources: mockDataError,
    }
}
