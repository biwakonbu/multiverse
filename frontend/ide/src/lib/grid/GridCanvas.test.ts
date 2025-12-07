/**
 * GridCanvas ズーム機能テスト
 *
 * CSS zoom プロパティを使用したテキスト品質向上の検証
 *
 * NOTE: JSDOM は CSS zoom プロパティをサポートしていないため、
 * style 属性から zoom を読み取ることはできません。
 * 実際のズーム品質テストは Storybook で視覚的に確認してください。
 * このテストではズームレベルの表示と基本的な機能を検証します。
 */

import { render, screen, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import GridCanvasPreview from './GridCanvasPreview.svelte';

// Mock design-system module
vi.mock('../../design-system', () => ({
  gridToCanvas: vi.fn((col: number, row: number) => ({
    x: col * 200,
    y: row * 140,
  })),
  zoom: { min: 0.25, max: 3, step: 0.1, default: 1, wheelFactor: 0.1 },
}));

describe('GridCanvas zoom functionality', () => {
  const sampleNodes = [
    {
      id: 'task-1',
      title: 'API設計',
      status: 'SUCCEEDED' as const,
      poolId: 'codegen',
      col: 0,
      row: 0,
    },
    {
      id: 'task-2',
      title: 'データベース設計',
      status: 'RUNNING' as const,
      poolId: 'codegen',
      col: 1,
      row: 0,
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('zoom indicator display', () => {
    it('shows 100% at default zoom level', () => {
      render(GridCanvasPreview, {
        props: {
          nodes: sampleNodes,
          zoom: 1,
          panX: 0,
          panY: 0,
          selectedId: null,
        },
      });

      expect(screen.getByText('100%')).toBeInTheDocument();
    });

    it('shows 200% at high zoom level', () => {
      render(GridCanvasPreview, {
        props: {
          nodes: sampleNodes,
          zoom: 2,
          panX: 0,
          panY: 0,
          selectedId: null,
        },
      });

      expect(screen.getByText('200%')).toBeInTheDocument();
    });

    it('shows 75% for fractional zoom', () => {
      render(GridCanvasPreview, {
        props: {
          nodes: sampleNodes,
          zoom: 0.75,
          panX: 0,
          panY: 0,
          selectedId: null,
        },
      });

      expect(screen.getByText('75%')).toBeInTheDocument();
    });

    it('shows 300% at maximum zoom level', () => {
      render(GridCanvasPreview, {
        props: {
          nodes: sampleNodes,
          zoom: 3,
          panX: 0,
          panY: 0,
          selectedId: null,
        },
      });

      expect(screen.getByText('300%')).toBeInTheDocument();
    });

    it('shows 25% at minimum zoom level', () => {
      render(GridCanvasPreview, {
        props: {
          nodes: sampleNodes,
          zoom: 0.25,
          panX: 0,
          panY: 0,
          selectedId: null,
        },
      });

      expect(screen.getByText('25%')).toBeInTheDocument();
    });
  });

  describe('nodes layer rendering', () => {
    it('renders nodes-layer element', () => {
      const { container } = render(GridCanvasPreview, {
        props: {
          nodes: sampleNodes,
          zoom: 2,
          panX: 50,
          panY: 50,
          selectedId: null,
        },
      });

      const nodesLayer = container.querySelector('.nodes-layer');
      expect(nodesLayer).toBeInTheDocument();
    });

    it('applies translate transform for panning', () => {
      const { container } = render(GridCanvasPreview, {
        props: {
          nodes: sampleNodes,
          zoom: 1,
          panX: 100,
          panY: 200,
          selectedId: null,
        },
      });

      const nodesLayer = container.querySelector('.nodes-layer');
      const style = nodesLayer?.getAttribute('style');
      // translate は適用されている
      expect(style).toContain('translate(100px, 200px)');
    });

    it('does NOT use scale transform (uses zoom property instead)', () => {
      const { container } = render(GridCanvasPreview, {
        props: {
          nodes: sampleNodes,
          zoom: 2,
          panX: 0,
          panY: 0,
          selectedId: null,
        },
      });

      const nodesLayer = container.querySelector('.nodes-layer');
      const style = nodesLayer?.getAttribute('style');
      // scale() が含まれていないことを確認（重要）
      // CSS zoom は JSDOM でサポートされていないため表示されないが、
      // scale が使われていないことは確認できる
      expect(style).not.toContain('scale(');
    });

    it('renders correct number of nodes', () => {
      const { container } = render(GridCanvasPreview, {
        props: {
          nodes: sampleNodes,
          zoom: 1.5,
          panX: 0,
          panY: 0,
          selectedId: null,
        },
      });

      const nodes = container.querySelectorAll('.node');
      expect(nodes.length).toBe(2);
    });
  });

  describe('node selection', () => {
    it('allows node selection via click', async () => {
      const { container } = render(GridCanvasPreview, {
        props: {
          nodes: sampleNodes,
          zoom: 1,
          panX: 0,
          panY: 0,
          selectedId: null,
        },
      });

      const nodeWrapper = container.querySelector('.nodes-layer > div');
      expect(nodeWrapper).toBeInTheDocument();

      if (nodeWrapper) {
        await fireEvent.click(nodeWrapper);
      }

      // 選択状態になることを確認（コンポーネント内部の状態変更）
      const selectedNode = container.querySelector('.node.selected');
      expect(selectedNode).toBeInTheDocument();
    });
  });
});

