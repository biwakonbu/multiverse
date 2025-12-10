import type { Meta, StoryObj } from '@storybook/svelte';
import DraggableWindow from './DraggableWindow.svelte';
import { html } from '@storybook/svelte';

const meta = {
  title: 'Components/UI/Window/DraggableWindow',
  component: DraggableWindow,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen', 
    docs: {
        description: {
            component: 'A floating, draggable, and resizable window component. Children are rendered inside the content area. NOTE: This component uses fixed positioning, so it may appear outside the typical story preview area if not carefully positioned.'
        }
    }
  },
  args: {
    title: 'Draggable Window',
    initialPosition: { x: 100, y: 100 },
    initialSize: { width: 400, height: 300 },
    id: 'demo-window',
  },
  argTypes: {
    children: { control: 'text' },
  }
} satisfies Meta<DraggableWindow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
   args: {
     title: 'Standard Window',
     children: () => html`<div style="padding: 20px;">This is the window content.</div>`,
   }
};

export const WithControls: Story = {
    args: {
        title: 'Window with Controls',
        controls: { close: true },
        children: () => html`<div style="padding: 20px;">Try closing me (check actions).</div>`,
    }
};
