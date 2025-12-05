import type { Meta, StoryObj } from '@storybook/svelte';
import Button from './Button.svelte';

const meta = {
  title: 'Design System/Button',
  component: Button,
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: { type: 'select' },
      options: ['primary', 'secondary', 'ghost', 'danger'],
      description: 'ボタンのスタイルバリアント'
    },
    size: {
      control: { type: 'select' },
      options: ['small', 'medium', 'large'],
      description: 'ボタンのサイズ'
    },
    disabled: {
      control: 'boolean',
      description: '無効状態'
    },
    label: {
      control: 'text',
      description: 'ボタンのラベル'
    }
  },
  args: {
    label: 'Button'
  }
} satisfies Meta<Button>;

export default meta;
type Story = StoryObj<typeof meta>;

// Primary ボタン（デフォルト）
export const Primary: Story = {
  args: {
    variant: 'primary',
    label: 'Primary Button'
  }
};

// Secondary ボタン
export const Secondary: Story = {
  args: {
    variant: 'secondary',
    label: 'Secondary Button'
  }
};

// Ghost ボタン
export const Ghost: Story = {
  args: {
    variant: 'ghost',
    label: 'Ghost Button'
  }
};

// Danger ボタン
export const Danger: Story = {
  args: {
    variant: 'danger',
    label: 'Danger Button'
  }
};

// サイズバリエーション
export const Small: Story = {
  args: {
    variant: 'primary',
    size: 'small',
    label: 'Small'
  }
};

export const Medium: Story = {
  args: {
    variant: 'primary',
    size: 'medium',
    label: 'Medium'
  }
};

export const Large: Story = {
  args: {
    variant: 'primary',
    size: 'large',
    label: 'Large'
  }
};

// 無効状態
export const Disabled: Story = {
  args: {
    variant: 'primary',
    disabled: true,
    label: 'Disabled Button'
  }
};

