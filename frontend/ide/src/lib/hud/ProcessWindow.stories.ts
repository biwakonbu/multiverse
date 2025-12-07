import type { Meta, StoryObj } from '@storybook/svelte';
import ProcessWindow from './ProcessWindow.svelte';
import { windowStore } from '../../stores/windowStore';
import type { ResourceNode } from './types';

const sampleResources: ResourceNode[] = [
  {
    id: 'worker-1',
    name: 'Docker Worker',
    type: 'WORKER',
    status: 'RUNNING',
    children: [
        { id: 'c-1', name: 'python-sandbox', type: 'CONTAINER', status: 'RUNNING', detail: 'CPU: 12%' },
        { id: 'c-2', name: 'node-sandbox', type: 'CONTAINER', status: 'IDLE', detail: 'Waiting for task' }
    ]
  },
  {
      id: 'meta-1',
      name: 'Meta Agent',
      type: 'META',
      status: 'THINKING',
      detail: 'Planning next steps...'
  }
];

const meta = {
  title: 'Features/HUD/ProcessWindow',
  component: ProcessWindow,
  tags: ['autodocs'],
  argTypes: {},
  parameters: {
      layout: 'fullscreen',
  },
  decorators: [
    (Story) => {
        windowStore.update((s: any) => ({ 
             ...s, 
             process: { ...s.process, isOpen: true, position: { x: 100, y: 100 } } 
        }));
        return Story();
    }
  ]
} as Meta<typeof ProcessWindow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
      resources: sampleResources
  } as any // Bypass strict Prop inference if failing
};
