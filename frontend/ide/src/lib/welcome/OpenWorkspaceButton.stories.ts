import type { Meta, StoryObj } from '@storybook/svelte';
import OpenWorkspaceButton from './OpenWorkspaceButton.svelte';

const meta = {
  title: 'Welcome/OpenWorkspaceButton',
  component: OpenWorkspaceButton,
  tags: ['autodocs'],
  argTypes: {
    loading: {
      control: 'boolean',
      description: 'ローディング状態',
    },
    disabled: {
      control: 'boolean',
      description: '無効状態',
    },
  },
  parameters: {
    backgrounds: {
      default: 'dark',
      values: [
        { name: 'dark', value: '#16181e' },
      ],
    },
  },
} as Meta<typeof OpenWorkspaceButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    loading: false,
    disabled: false,
  },
};

export const Loading: Story = {
  args: {
    loading: true,
    disabled: false,
  },
};

export const Disabled: Story = {
  args: {
    loading: false,
    disabled: true,
  },
};
