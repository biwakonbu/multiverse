import type { Meta, StoryObj } from '@storybook/svelte';
import type { ComponentProps } from 'svelte';
import Box from './Box.svelte';

const meta = {
  title: 'Design System/Layout/Box',
  component: Box,
  tags: ['autodocs'],
  argTypes: {
    as: { control: 'text' },
    p: { control: 'text' },
    m: { control: 'text' },
    bg: { control: 'text' },
    color: { control: 'text' },
    radius: { control: 'text' },
  },
  args: {
    p: 'var(--mv-spacing-md)',
    bg: 'var(--mv-color-surface-primary)',
    color: 'var(--mv-color-text-primary)',
    radius: 'var(--mv-radius-md)',
  }
} as Meta<typeof Box>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args: ComponentProps<typeof Box>) => ({
    Component: Box,
    props: args,
    // Slot content needs to be handled via template or specific render function in Svelte Storybook, 
    // but simplified text slot works for basic check
  }),
  parameters: {
    slots: {
      default: 'Box Content'
    }
  }
};

export const WithBorder: Story = {
  args: {
    border: '1px solid var(--mv-color-border-default)',
    bg: 'var(--mv-color-surface-secondary)',
  }
};
