import type { Meta, StoryObj } from '@storybook/svelte';
import Button from './Button.svelte';

const meta = {
  title: 'Design System/Components/Button',
  component: Button,
  tags: ['autodocs'],
  argTypes: {
    variant: { 
      control: 'select', 
      options: ['primary', 'secondary', 'ghost', 'danger', 'crystal'] 
    },
    size: { 
      control: 'select', 
      options: ['small', 'medium', 'large'] 
    },
    disabled: { control: 'boolean' },
    loading: { control: 'boolean' },
    label: { control: 'text' },
  },
  args: {
    variant: 'primary',
    size: 'medium',
    label: 'Button',
  }
} as Meta<typeof Button>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Primary: Story = {};

export const Secondary: Story = {
  args: {
    variant: 'secondary',
  }
};

export const Ghost: Story = {
  args: {
    variant: 'ghost',
  }
};

export const Danger: Story = {
  args: {
    variant: 'danger',
  }
};

export const Crystal: Story = {
  args: {
    variant: 'crystal',
  },
  globals: {
    backgrounds: {
      value: "multiverse-app"
    }
  }
};

export const Loading: Story = {
  args: {
    loading: true,
  }
};
