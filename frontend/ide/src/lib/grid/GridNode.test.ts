
import { render, screen } from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import GridNode from './GridNode.svelte';

// Mock dependencies
vi.mock('../../stores', () => ({
  selectedTaskId: {
    subscribe: vi.fn(run => {
        run(null);
        return () => {};
    }),
    select: vi.fn()
  }
}));

vi.mock('../../design-system', () => ({
  gridToCanvas: vi.fn(() => ({ x: 0, y: 0 })),
  zoom: { min: 0.1, max: 2, step: 0.1 }
}));

describe('GridNode', () => {
  const mockTask = {
    id: '1',
    title: 'Test Task',
    status: 'PENDING' as const,
    phaseName: '概念設計' as const,
    dependencies: [],
    poolId: 'default',
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z'
  };

  it('renders title without markdown backticks', () => {
    const taskWithMarkdown = {
      ...mockTask,
      title: 'Fix `auth` module'
    };

    render(GridNode, { 
      props: { 
        task: taskWithMarkdown, 
        col: 0, 
        row: 0, 
        zoomLevel: 1 
      } 
    });

    // The component currently renders raw text, so this test IS EXPECTED TO FAIL initially.
    // If it passes immediately, then something is wrong with the test or the assumption.
    const titleEl = screen.getByTitle('Fix `auth` module');
    expect(titleEl.textContent?.trim()).toBe('Fix auth module');
  });

  it('renders title without triple backticks', () => {
     const taskWithBlock = {
      ...mockTask,
      title: '```Clean up code```'
    };

    render(GridNode, { 
      props: { 
        task: taskWithBlock, 
        col: 0, 
        row: 0, 
        zoomLevel: 1 
      } 
    });
    
    const titleEl = screen.getByTitle('```Clean up code```');
    expect(titleEl.textContent?.trim()).toBe('Clean up code');
  });
});
