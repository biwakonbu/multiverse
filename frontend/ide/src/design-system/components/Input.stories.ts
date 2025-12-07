import type { Meta, StoryObj } from '@storybook/svelte-vite';
import Input from './Input.svelte';

const meta = {
  title: 'Design System/Components/Input',
  component: Input,
  tags: ['autodocs'],
  argTypes: {
    type: { 
      control: 'select', 
      options: ['text', 'password', 'search', 'email'] 
    },
    placeholder: { control: 'text' },
    label: { control: 'text' },
    error: { control: 'text' },
    disabled: { control: 'boolean' },
    value: { control: 'text' },
  },
  args: {
    type: 'text',
    placeholder: 'Enter text...',
    label: 'Label',
  }
} satisfies Meta<Input>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithError: Story = {
  args: {
    value: 'invalid-input',
    error: 'This field is required',
  }
};

export const Disabled: Story = {
  args: {
    disabled: true,
  }
};

export const Password: Story = {
  args: {
    type: 'password',
    label: 'Password',
    placeholder: '••••••••',
  }
};
