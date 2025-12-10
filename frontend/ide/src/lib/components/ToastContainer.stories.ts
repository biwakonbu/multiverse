import type { Meta, StoryObj } from '@storybook/svelte';
import ToastContainer from './ToastContainer.svelte';
import { toasts } from '../../stores/toastStore';

const meta = {
  title: 'Components/ToastContainer',
  component: ToastContainer,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
  },
} satisfies Meta<ToastContainer>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  play: async () => {
    // Reset toasts
    toasts.set([]);
    
    // Add demo toasts
    toasts.add({
      type: 'info',
      message: 'This is an info toast',
    });
    
    setTimeout(() => {
      toasts.add({
        type: 'success',
        message: 'This is a success toast',
      });
    }, 500);

    setTimeout(() => {
        toasts.add({
          type: 'warning',
          message: 'This is a warning toast',
        });
      }, 1000);

    setTimeout(() => {
      toasts.add({
        type: 'error',
        message: 'This is an error toast',
      });
    }, 1500);
  }
};

export const WithAction: Story = {
    play: async () => {
        toasts.set([]);
        toasts.add({
            type: 'info',
            message: 'Toast with action',
            action: {
                label: 'Undo',
                onClick: () => console.log('Undo clicked')
            }
        });
    }
}
