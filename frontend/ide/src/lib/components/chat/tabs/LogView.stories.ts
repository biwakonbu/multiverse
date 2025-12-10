import type { Meta, StoryObj } from '@storybook/svelte';
import LogView from './LogView.svelte';
import { chatLog } from '../../../../stores/chat';

const meta = {
  title: 'Components/Chat/Tabs/LogView',
  component: LogView,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
  },
} satisfies Meta<LogView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  play: async () => {
    chatLog.set([
      { timestamp: Date.now(), step: 'Planning', message: 'Analyzing requirements...' },
      { timestamp: Date.now() + 1000, step: 'Execution', message: 'Running tool: list_dir' },
      { timestamp: Date.now() + 2000, step: 'Execution', message: 'Found 5 files.' },
      { timestamp: Date.now() + 3000, step: 'Verification', message: 'Verifying changes...' },
    ]);
  },
};

export const Empty: Story = {
    play: async () => {
        chatLog.set([]);
    }
};
