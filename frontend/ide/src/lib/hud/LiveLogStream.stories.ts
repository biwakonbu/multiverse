
import type { Meta, StoryObj } from '@storybook/svelte';
import LiveLogStream from './LiveLogStream.svelte';

const meta = {
  title: 'HUD/LiveLogStream',
  component: LiveLogStream,
  tags: ['autodocs'],
  parameters: {
      backgrounds: {
          default: 'dark',
      }
  }
} as Meta<typeof LiveLogStream>;

export default meta;
type Story = StoryObj<typeof meta>;

const mockLogs = [
    { id: '1', taskId: 't1', stream: 'stdout', line: 'Initializing agent runner...', timestamp: '2023-01-01T12:00:00.000Z' },
    { id: '2', taskId: 't1', stream: 'stdout', line: 'Loading configuration', timestamp: '2023-01-01T12:00:01.000Z' },
    { id: '3', taskId: 't1', stream: 'stderr', line: 'Warning: Cache miss', timestamp: '2023-01-01T12:00:01.500Z' },
    { id: '4', taskId: 't1', stream: 'stdout', line: 'Docker container started: vigorous_turing', timestamp: '2023-01-01T12:00:02.000Z' },
    { id: '5', taskId: 't1', stream: 'stdout', line: 'Running: go test ./...', timestamp: '2023-01-01T12:00:02.500Z' },
] as any[]; // Type assertion to avoid store dependency issues in stories if strict

export const Default: Story = {
  args: {
    logs: [],
    height: '200px',
  },
};

export const WithLogs: Story = {
  args: {
    logs: mockLogs,
    height: '300px',
  },
};

export const ManyLogs: Story = {
    args: {
        logs: Array.from({ length: 50 }).map((_, i) => ({
            id: i.toString(),
            taskId: 't1',
            stream: i % 5 === 0 ? 'stderr' : 'stdout',
            line: `Log entry number ${i} - ${i % 5 === 0 ? 'Error occurred' : 'Processing step...'}`,
            timestamp: new Date().toISOString()
        })) as any[],
        height: '300px'
    }
}
