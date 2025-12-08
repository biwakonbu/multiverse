import dagre from 'dagre';
import { Position, type Node, type Edge } from '@xyflow/svelte';
import type { Task } from '../../types';

const nodeWidth = 200;
const nodeHeight = 80; // Approximate height with header and content

export const getLayoutedElements = (
  nodes: Node[],
  edges: Edge[],
  direction = 'TB' // 'TB' (top to bottom) or 'LR' (left to right)
) => {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));

  const isHorizontal = direction === 'LR';
  dagreGraph.setGraph({ rankdir: direction });

  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
  });

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  dagre.layout(dagreGraph);

  const layoutedNodes = nodes.map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);
    
    // We are shifting the dagre node position (anchor=center center) to the top left
    // so it matches the React Flow node anchor point (top left).
    return {
      ...node,
      targetPosition: isHorizontal ? Position.Left : Position.Top,
      sourcePosition: isHorizontal ? Position.Right : Position.Bottom,
      position: {
        x: nodeWithPosition.x - nodeWidth / 2,
        y: nodeWithPosition.y - nodeHeight / 2,
      },
    };
  });

  return { nodes: layoutedNodes, edges };
};

export function convertTasksToFlowData(tasks: Task[]) {
  const nodes: Node[] = [];
  const edges: Edge[] = [];

  tasks.forEach((task) => {
    // Node
    nodes.push({
      id: task.id,
      type: 'task', // Custom node type
      position: { x: 0, y: 0 }, // Initial position, will be calculated by dagre
      data: { task },
    });

    // Edges (Dependencies)
    task.dependencies?.forEach((depId: string) => {
      edges.push({
        id: `e${depId}-${task.id}`,
        source: depId,
        target: task.id,
        type: 'dependency', // Custom edge type
        animated: true,
      });
    });
  });

  return { nodes, edges };
}
