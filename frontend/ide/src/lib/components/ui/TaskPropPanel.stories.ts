
import type { Meta, StoryObj } from '@storybook/svelte';
import TaskPropPanelPreview from './TaskPropPanelPreview.svelte';
import type { Task } from '../../../types';

const meta = {
  title: 'UI/TaskPropPanel',
  component: TaskPropPanelPreview,
  tags: ['autodocs'],
  argTypes: {
    task: {
        control: 'object',
        description: '表示するタスクデータ (Mock)'
    }
  },
} satisfies Meta<typeof TaskPropPanelPreview>;

export default meta;
type Story = StoryObj<typeof meta>;

const baseTask: Task = {
    id: 'task-1',
    title: 'Sample Task',
    status: 'RUNNING',
    poolId: 'default',
    phaseName: '実装',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    description: 'This is a sample task description.',
    dependencies: [],
};

export const Default: Story = {
  args: {
    task: baseTask
  }
};

export const WithSuggestedImpl: Story = {
    args: {
        task: {
            ...baseTask,
            title: 'Task with Implementation Plan',
            suggestedImpl: {
                language: 'typescript',
                filePaths: [
                    'src/lib/components/ui/TaskPropPanel.svelte',
                    'src/types/index.ts'
                ],
                constraints: [
                    'Use Svelte 5 runes',
                    'Ensure type safety'
                ]
            }
        }
    }
};

export const WithArtifacts: Story = {
    args: {
        task: {
            ...baseTask,
            title: 'Task with Artifacts',
            status: 'SUCCEEDED',
            artifacts: {
                files: ['frontend/dist/bundle.js'],
                logs: ['build.log']
            }
        }
    }
};

export const FullFeatures: Story = {
    args: {
        task: {
            ...baseTask,
            title: 'Full Featured Task',
            status: 'SUCCEEDED',
            suggestedImpl: {
                language: 'go',
                filePaths: ['cmd/main.go'],
                constraints: ['No panics']
            },
            artifacts: {
                files: ['bin/app'],
                logs: ['compile.log', 'test.log']
            }
        }
    }
};
