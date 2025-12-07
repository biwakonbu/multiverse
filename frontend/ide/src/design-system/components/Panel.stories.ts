import type { Meta, StoryObj } from '@storybook/svelte';
import Panel from './Panel.svelte';

const meta = {
  title: 'Design System/Layout/Panel',
  component: Panel,
  tags: ['autodocs'],
  argTypes: {
    variant: { 
      control: 'select', 
      options: ['default', 'glass', 'glass-strong', 'outlined'] 
    },
    padding: { 
      control: 'select', 
      options: ['none', 'sm', 'md', 'lg'] 
    },
    radius: { 
      control: 'select', 
      options: ['none', 'sm', 'md', 'lg', 'full'] 
    },
    hover: { control: 'boolean' },
    glow: { control: 'boolean' },
  },
  args: {
    variant: 'glass',
    padding: 'md',
    radius: 'md',
    hover: false,
    glow: false,
  }
} satisfies Meta<Panel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Glass: Story = {
  parameters: {
    slots: {
      default: '<div style="color: white;">Phantom Glass Panel Content</div>'
    }
  }
};

export const GlassStrong: Story = {
  args: {
    variant: 'glass-strong',
  },
  parameters: {
    slots: {
      default: '<div style="color: white; font-weight: bold;">Strong Glass Panel</div>'
    }
  }
};

export const HoverEffects: Story = {
  args: {
    hover: true,
  },
  parameters: {
    slots: {
      default: '<div style="color: white;">Hover me!</div>'
    }
  }
};

export const WithGlow: Story = {
  args: {
    glow: true,
  },
  parameters: {
    slots: {
      default: '<div style="color: white;">Glowing Panel</div>'
    }
  }
};
