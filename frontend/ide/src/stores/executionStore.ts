/**
 * 実行状態管理ストア
 *
 * ExecutionOrchestrator の状態をフロントエンドで管理
 * Wails バインディングが生成された後、実際の API に接続
 */

import { writable } from 'svelte/store';

export type ExecutionState = 'IDLE' | 'RUNNING' | 'PAUSED';

export const executionState = writable<ExecutionState>('IDLE');

// Wails イベントリスナー初期化（Wails バインディング生成後に有効化）
export function initExecutionEvents(): void {
    // TODO: Wails バインディング生成後、以下のコードを有効化
    // runtime.EventsOn('execution:stateChange', (event: { newState: ExecutionState }) => {
    //     executionState.set(event.newState);
    // });
    console.log('[ExecutionStore] initExecutionEvents called (stub)');
}

// 実行開始（スタブ実装）
export async function startExecution(): Promise<void> {
    // TODO: Wails バインディング生成後、StartExecution() を呼び出し
    console.log('[ExecutionStore] startExecution called (stub)');
    executionState.set('RUNNING');
}

// 実行一時停止（スタブ実装）
export async function pauseExecution(): Promise<void> {
    // TODO: Wails バインディング生成後、PauseExecution() を呼び出し
    console.log('[ExecutionStore] pauseExecution called (stub)');
    executionState.set('PAUSED');
}

// 実行再開（スタブ実装）
export async function resumeExecution(): Promise<void> {
    // TODO: Wails バインディング生成後、ResumeExecution() を呼び出し
    console.log('[ExecutionStore] resumeExecution called (stub)');
    executionState.set('RUNNING');
}

// 実行停止（スタブ実装）
export async function stopExecution(): Promise<void> {
    // TODO: Wails バインディング生成後、StopExecution() を呼び出し
    console.log('[ExecutionStore] stopExecution called (stub)');
    executionState.set('IDLE');
}
