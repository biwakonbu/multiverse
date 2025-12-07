import type { Meta, StoryObj } from '@storybook/svelte';
import type { ComponentProps } from 'svelte';
import Tooltip from './Tooltip.svelte';

const meta = {
  title: 'Design System/Feedback/Tooltip',
  component: Tooltip,
  tags: ['autodocs'],
  argTypes: {
    content: { control: 'text' },
    position: {
      control: 'select',
      options: ['top', 'bottom', 'left', 'right']
    },
  },
  args: {
    content: 'This is a tooltip',
    position: 'top',
  }
} as Meta<typeof Tooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

// Wrapper for demonstration since Svelte slots in Storybook are strings
export const Default: Story = {
  render: (args: ComponentProps<typeof Tooltip>) => ({
    Component: Tooltip,
    props: args,
  }),
  parameters: {
     slots: {
      default: '<button style="padding: 10px;">Hover Me</button>'
    }
  }
};

export const Right: Story = {
  args: {
    position: 'right',
  },
  parameters: {
     slots: {
      default: '<button style="padding: 10px;">Hover Me</button>'
    }
  }
};
