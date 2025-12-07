
import type { Meta, StoryObj } from '@storybook/svelte';
import ProcessHUD from './ProcessHUD.svelte';
import type { ResourceNode } from '../../stores/processStore';

const meta = {
  title: 'HUD/ProcessHUD',
  component: ProcessHUD,
  tags: ['autodocs'],
  parameters: {
      layout: 'fullscreen',
      backgrounds: {
          default: 'dark',
      }
  }
} as Meta<typeof ProcessHUD>;

export default meta;
type Story = StoryObj<typeof meta>;

const mockResources: ResourceNode[] = [
    {
        id: 'sys',
        name: 'Multiverse Orchestrator',
        type: 'ORCHESTRATOR',
        status: 'RUNNING',
        detail: 'Active Task: Implement Feature X',
        expanded: true,
        children: [
            {
                id: 'meta-1',
                name: 'Meta-Agent',
                type: 'META',
                status: 'THINKING',
                detail: 'Planning next steps...',
                expanded: true,
                children: [
                    {
                        id: 'worker-1',
                        name: 'Worker: Codex',
                        type: 'WORKER',
                        status: 'RUNNING',
                        detail: 'executing: go test -v ./...',
                        expanded: true,
                        children: [
                             {
                                 id: 'container-1',
                                 name: 'Container',
                                 type: 'CONTAINER',
                                 status: 'RUNNING',
                                 detail: 'Image: alpine:latest'
                             }
                        ]
                    }
                ]
            }
        ]
    }
];

export const Idle: Story = {
  args: {
    executionState: 'IDLE',
    resources: [],
  },
};

export const Running: Story = {
  args: {
    executionState: 'RUNNING',
    resources: mockResources,
    activeTaskTitle: 'Implementing Feature X',
  },
};

export const Paused: Story = {
  args: {
    executionState: 'PAUSED',
    resources: mockResources, // Keep resources visible while paused
  },
};
