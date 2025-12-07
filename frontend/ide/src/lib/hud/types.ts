/**
 * HUD コンポーネント用の型定義
 * Storybook 環境での wailsjs 依存を避けるため、
 * processStore.ts からではなくこのファイルから型をインポートする
 */

export type ResourceType = 'META' | 'WORKER' | 'CONTAINER' | 'ORCHESTRATOR';
export type ResourceStatus = 'IDLE' | 'RUNNING' | 'THINKING' | 'PAUSED' | 'ERROR' | 'DONE';

export interface ResourceNode {
  id: string;
  name: string;
  type: ResourceType;
  status: ResourceStatus;
  detail?: string;
  children?: ResourceNode[];
  expanded?: boolean;
}
