import type { Meta, StoryObj } from '@storybook/svelte';
import MockMainView from './MockMainView.svelte';

const meta = {
  title: 'Features/Chat/FloatingChatWindow',
  component: MockMainView,
  tags: ['autodocs'],
  argTypes: {},
  parameters: {
    layout: 'fullscreen',
    docs: {
      description: {
        component: 'フローティングチャットウィンドウ。ドラッグ可能なウィンドウ内にチャットUIを表示します。',
      },
    },
  },
} as Meta<typeof MockMainView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  parameters: {
    docs: {
      description: {
        story: 'デフォルト状態のフローティングチャットウィンドウ。',
      },
    },
  },
};
