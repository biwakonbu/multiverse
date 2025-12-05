import type { Meta, StoryObj } from '@storybook/svelte';
import Input from './Input.svelte';

const meta = {
  title: 'Design System/Input',
  component: Input,
  tags: ['autodocs'],
  argTypes: {
    type: {
      control: { type: 'select' },
      options: ['text', 'password', 'email', 'number', 'search'],
      description: '入力タイプ'
    },
    value: {
      control: 'text',
      description: '入力値'
    },
    placeholder: {
      control: 'text',
      description: 'プレースホルダー'
    },
    label: {
      control: 'text',
      description: 'ラベル'
    },
    disabled: {
      control: 'boolean',
      description: '無効状態'
    },
    error: {
      control: 'boolean',
      description: 'エラー状態'
    },
    errorMessage: {
      control: 'text',
      description: 'エラーメッセージ'
    },
    size: {
      control: { type: 'select' },
      options: ['small', 'medium', 'large'],
      description: 'サイズ'
    }
  }
} satisfies Meta<Input>;

export default meta;
type Story = StoryObj<typeof meta>;

// 基本
export const Default: Story = {
  args: {
    placeholder: 'Enter text...'
  }
};

// ラベル付き
export const WithLabel: Story = {
  args: {
    label: 'Task Title',
    placeholder: 'Enter task title...'
  }
};

// サイズ
export const Small: Story = {
  args: {
    size: 'small',
    placeholder: 'Small input'
  }
};

export const Medium: Story = {
  args: {
    size: 'medium',
    placeholder: 'Medium input'
  }
};

export const Large: Story = {
  args: {
    size: 'large',
    placeholder: 'Large input'
  }
};

// 無効状態
export const Disabled: Story = {
  args: {
    label: 'Disabled Input',
    value: 'Cannot edit',
    disabled: true
  }
};

// エラー状態
export const Error: Story = {
  args: {
    label: 'Email',
    value: 'invalid-email',
    error: true,
    errorMessage: 'Please enter a valid email address'
  }
};

// パスワード
export const Password: Story = {
  args: {
    type: 'password',
    label: 'Password',
    placeholder: 'Enter password...'
  }
};

// 検索
export const Search: Story = {
  args: {
    type: 'search',
    placeholder: 'Search tasks...'
  }
};
