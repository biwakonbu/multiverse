import type { Meta, StoryObj } from '@storybook/svelte-vite';
import WelcomeHeader from './WelcomeHeader.svelte';

const meta = {
  title: 'Welcome/WelcomeHeader',
  component: WelcomeHeader,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
    backgrounds: {
      default: 'dark',
      values: [
        { name: 'dark', value: '#16181e' }
      ]
    },
    docs: {
      description: {
        component: 'ウェルカム画面のヘッダー部分。ロゴ、タイトル、サブタイトルを表示。'
      }
    }
  }
} satisfies Meta<WelcomeHeader>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
