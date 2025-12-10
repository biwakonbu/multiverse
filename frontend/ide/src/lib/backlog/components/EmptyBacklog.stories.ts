import type { Meta, StoryObj } from '@storybook/svelte';
import EmptyBacklog from './EmptyBacklog.svelte';

const meta = {
  title: 'Backlog/Components/EmptyBacklog',
  component: EmptyBacklog,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
} satisfies Meta<EmptyBacklog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
