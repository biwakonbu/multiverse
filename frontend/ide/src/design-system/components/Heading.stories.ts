import type { Meta, StoryObj } from '@storybook/svelte';
import Heading from './Heading.svelte';

const meta = {
  title: 'Design System/Typography/Heading',
  component: Heading,
  tags: ['autodocs'],
  argTypes: {
    level: { 
      control: { type: 'number', min: 1, max: 6 } 
    },
    variant: { 
      control: 'select', 
      options: ['default', 'gradient'] 
    },
  },
  args: {
    level: 1,
    variant: 'default',
  }
} satisfies Meta<Heading>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  parameters: {
    slots: {
      default: 'Multiverse IDE'
    }
  }
};

export const GradientBrand: Story = {
  args: {
    variant: 'gradient',
  },
  parameters: {
    slots: {
      default: 'Multiverse IDE'
    }
  }
};

export const Level2: Story = {
  args: {
    level: 2,
  },
  parameters: {
    slots: {
      default: 'Section Title'
    }
  }
};
