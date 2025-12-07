import type { Meta, StoryObj } from '@storybook/svelte';
import Card from './Card.svelte';

const meta = {
  title: 'Design System/Card',
  component: Card,
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: { type: 'select' },
      options: ['default', 'elevated', 'outlined'],
      description: 'カードのスタイルバリアント'
    },
    padding: {
      control: { type: 'select' },
      options: ['none', 'small', 'medium', 'large'],
      description: 'パディングサイズ'
    },
    selected: {
      control: 'boolean',
      description: '選択状態'
    },
    interactive: {
      control: 'boolean',
      description: 'インタラクティブ（ホバー効果）'
    }
  }
} as Meta<typeof Card>;

export default meta;
type Story = StoryObj<typeof meta>;

// バリアント
export const Default: Story = {
  args: {
    variant: 'default'
  }
};

export const Elevated: Story = {
  args: {
    variant: 'elevated'
  }
};

export const Outlined: Story = {
  args: {
    variant: 'outlined'
  }
};

// パディング
export const PaddingSmall: Story = {
  args: {
    variant: 'default',
    padding: 'small'
  }
};

export const PaddingLarge: Story = {
  args: {
    variant: 'default',
    padding: 'large'
  }
};

// 選択状態
export const Selected: Story = {
  args: {
    variant: 'default',
    selected: true
  }
};

// インタラクティブ
export const Interactive: Story = {
  args: {
    variant: 'default',
    interactive: true
  }
};
