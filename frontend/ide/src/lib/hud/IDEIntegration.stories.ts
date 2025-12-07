
import type { Meta, StoryObj } from '@storybook/svelte';
import IDEMockLayout from './IDEMockLayout.svelte';

const meta = {
  title: 'HUD/Integration/IDE Overlay',
  component: IDEMockLayout,
  parameters: {
      layout: 'fullscreen',
  }
} as Meta<typeof IDEMockLayout>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Demo: Story = {
    args: {
        executionState: 'RUNNING'
    }
};
