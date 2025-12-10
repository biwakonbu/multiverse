import type { Meta, StoryObj } from '@storybook/svelte';
import ZoomControls from './ZoomControls.svelte';
import { fn } from '@storybook/test';

const meta = {
  title: 'WBS/ZoomControls',
  component: ZoomControls,
  tags: ['autodocs'],
  args: {
    onzoomout: fn(),
    onzoomin: fn(),
    onreset: fn(),
  },
  parameters: {
      layout: 'centered',
  }    
} satisfies Meta<ZoomControls>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    scale: 1,
  },
};

export const ZoomedIn: Story = {
  args: {
    scale: 1.5,
  },
};

export const ZoomedOut: Story = {
  args: {
    scale: 0.5,
  },
};
