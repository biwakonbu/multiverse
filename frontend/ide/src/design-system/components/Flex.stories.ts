import type { Meta, StoryObj } from '@storybook/svelte-vite';
import Flex from './Flex.svelte';

const meta = {
  title: 'Design System/Layout/Flex',
  component: Flex,
  tags: ['autodocs'],
  argTypes: {
    direction: { control: 'select', options: ['row', 'column'] },
    align: { control: 'select', options: ['start', 'center', 'end', 'stretch'] },
    justify: { control: 'select', options: ['start', 'center', 'end', 'between'] },
    gap: { control: 'text' },
  },
  args: {
    gap: 'var(--mv-spacing-md)',
    p: 'var(--mv-spacing-md)',
    bg: 'var(--mv-color-surface-secondary)',
  }
} satisfies Meta<Flex>;

export default meta;
type Story = StoryObj<typeof meta>;

// Helper to create children html string for slots in storybook is tricky in Svelte without creating a wrapper component.
// We will rely on simple text or assumption that checking layout needs visual inspection.
// For automated testing or better stories, component composition is usually done via a wrapper .svelte file in stories.
// But we can try to inject HTML if the component supports it, or valid Svelte slots.
// Since we can't easily pass sub-components in args for slot, we will use a basic usage example.

export const Row: Story = {
  render: (args) => ({
    Component: Flex,
    props: args,
  }),
  parameters: {
    slots: {
      default: `
        <div style="background: var(--mv-color-status-running-text); padding: 8px; color: black;">Item 1</div>
        <div style="background: var(--mv-color-status-running-text); padding: 8px; color: black;">Item 2</div>
        <div style="background: var(--mv-color-status-running-text); padding: 8px; color: black;">Item 3</div>
      `
    }
  }
};

export const Column: Story = {
  args: {
    direction: 'column',
  },
  parameters: {
     slots: {
      default: `
        <div style="background: var(--mv-color-status-running-text); padding: 8px; color: black;">Item 1</div>
        <div style="background: var(--mv-color-status-running-text); padding: 8px; color: black;">Item 2</div>
        <div style="background: var(--mv-color-status-running-text); padding: 8px; color: black;">Item 3</div>
      `
    }
  }
};
