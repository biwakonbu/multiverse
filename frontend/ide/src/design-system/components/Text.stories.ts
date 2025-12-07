import type { Meta, StoryObj } from '@storybook/svelte-vite';
import Text from './Text.svelte';

const meta = {
  title: 'Design System/Typography/Text',
  component: Text,
  tags: ['autodocs'],
  argTypes: {
    variant: { 
      control: 'select', 
      options: ['primary', 'secondary', 'muted', 'disabled', 'success', 'warning', 'error', 'info'] 
    },
    size: { 
      control: 'select', 
      options: ['xs', 'sm', 'md', 'lg', 'xl'] 
    },
    weight: { 
      control: 'select', 
      options: ['normal', 'medium', 'semibold', 'bold'] 
    },
    mono: { control: 'boolean' },
    glow: { control: 'boolean' },
    as: { control: 'text' },
  },
  args: {
    variant: 'primary',
    size: 'md',
    weight: 'normal',
    mono: false,
    glow: false,
    as: 'p',
  }
} satisfies Meta<Text>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  parameters: {
    slots: {
      default: 'The quick brown fox jumps over the lazy dog.'
    }
  }
};

export const Secondary: Story = {
  args: {
    variant: 'secondary',
  },
  parameters: {
    slots: {
      default: 'This is secondary text information.'
    }
  }
};

export const Mono: Story = {
  args: {
    mono: true,
  },
  parameters: {
    slots: {
      default: 'Consolas, Monaco, "Andale Mono", "Ubuntu Mono", monospace'
    }
  }
};

export const GlowingSuccess: Story = {
  args: {
    variant: 'success',
    glow: true,
    weight: 'bold',
  },
  parameters: {
    slots: {
      default: 'OPERATION SUCCESSFUL'
    }
  }
};
