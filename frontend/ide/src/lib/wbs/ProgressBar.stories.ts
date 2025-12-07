import type { Meta, StoryObj } from '@storybook/svelte-vite';
import ProgressBar from './ProgressBar.svelte';

const meta = {
  title: 'WBS/ProgressBar',
  component: ProgressBar,
  tags: ['autodocs'],
  argTypes: {
    percentage: { control: { type: 'range', min: 0, max: 100, step: 1 } },
    size: { control: { type: 'select' }, options: ['sm', 'md', 'mini'] },
    className: { control: 'text' },
  },
  parameters: {
    layout: 'centered',
    backgrounds: {
      default: 'dark',
    }
  },
} satisfies Meta<ProgressBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    percentage: 50,
    size: 'sm',
  },
};

export const Large: Story = {
  args: {
    percentage: 75,
    size: 'md',
  },
};
