import type { Meta, StoryObj } from '@storybook/svelte-vite';
import FloatingChatWindow from './FloatingChatWindow.svelte';
import MockMainView from './MockMainView.svelte';

const meta = {
  title: 'Features/Chat/FloatingChatWindow',
  component: FloatingChatWindow,
  tags: ['autodocs'],
  argTypes: {},
  parameters: {
      layout: 'centered',
  }
} satisfies Meta<FloatingChatWindow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Standalone: Story = {
  args: {
    initialPosition: { x: 0, y: 0 }, // Relative to story container
  },
  parameters: {
      layout: 'centered', // Show in center to focus on component
  }
};

export const InLayout: Story = {
    render: () => ({
        Component: MockMainView,
        props: {}
    }),
    args: {},
    parameters: {
        layout: 'fullscreen'
    }
}
