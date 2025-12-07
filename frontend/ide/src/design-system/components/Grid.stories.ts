import type { Meta, StoryObj } from '@storybook/svelte-vite';
import Grid from './Grid.svelte';

const meta = {
  title: 'Design System/Layout/Grid',
  component: Grid,
  tags: ['autodocs'],
  argTypes: {
    columns: { control: 'text' },
    rows: { control: 'text' },
    gap: { control: 'text' },
    align: { control: 'select', options: ['start', 'center', 'end', 'stretch'] },
    justify: { control: 'select', options: ['start', 'center', 'end', 'stretch'] },
  },
  args: {
    columns: 'repeat(3, 1fr)',
    gap: 'var(--mv-spacing-md)',
    p: 'var(--mv-spacing-md)',
  }
} satisfies Meta<Grid>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    bg: 'var(--mv-color-surface-secondary)',
  },
  parameters: {
    slots: {
      default: `
        <div style="background: var(--mv-primitive-frost-1); padding: 16px;">1</div>
        <div style="background: var(--mv-primitive-frost-2); padding: 16px;">2</div>
        <div style="background: var(--mv-primitive-frost-3); padding: 16px;">3</div>
        <div style="background: var(--mv-primitive-frost-3); padding: 16px;">4</div>
        <div style="background: var(--mv-primitive-frost-2); padding: 16px;">5</div>
        <div style="background: var(--mv-primitive-frost-1); padding: 16px;">6</div>
      `
    }
  }
};

export const AutoFit: Story = {
  args: {
    columns: 'repeat(auto-fit, minmax(100px, 1fr))',
    bg: 'var(--mv-color-surface-secondary)',
  },
  parameters: {
    slots: {
      default: `
        <div style="background: var(--mv-primitive-frost-1); padding: 16px;">1</div>
        <div style="background: var(--mv-primitive-frost-2); padding: 16px;">2</div>
        <div style="background: var(--mv-primitive-frost-3); padding: 16px;">3</div>
      `
    }
  }
};
