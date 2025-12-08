import type { Meta, StoryObj } from '@storybook/svelte';
import DependencyEdgeStory from './DependencyEdgeStory.svelte';

const meta = {
  title: 'Flow/Edges/DependencyEdge',
  component: DependencyEdgeStory,
  tags: ['autodocs'],
  argTypes: {
    satisfied: { control: 'boolean' },
    animated: { control: 'boolean' },
  },
} satisfies Meta<DependencyEdgeStory>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    satisfied: false,
    animated: false,
  },
};

export const Satisfied: Story = {
  args: {
    satisfied: true,
    animated: false,
  },
};

export const Animated: Story = {
  args: {
    satisfied: false,
    animated: true,
  },
};
