/**
 * 実行状態管理ストア
 *
 * ExecutionOrchestrator の状態をフロントエンドで管理
 * Wails バインディングが生成された後、実際の API に接続
 */

import { writable } from 'svelte/store';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { StartExecution, PauseExecution, ResumeExecution, StopExecution } from '../../wailsjs/go/main/App';

export type ExecutionState = 'IDLE' | 'RUNNING' | 'PAUSED';

export const executionState = writable<ExecutionState>('IDLE');

// Wails イベントリスナー初期化（Wails バインディング生成後に有効化）
export function initExecutionEvents(): void {
    EventsOn('execution:stateChange', (event: { newState: ExecutionState }) => {
        executionState.set(event.newState);
    });
}

// 実行開始
export async function startExecution(): Promise<void> {
    try {
        await StartExecution();
    } catch (e) {
        console.error('Failed to start execution:', e);
    }
}

// 実行一時停止
export async function pauseExecution(): Promise<void> {
    try {
        await PauseExecution();
    } catch (e) {
        console.error('Failed to pause execution:', e);
    }
}

// 実行再開
export async function resumeExecution(): Promise<void> {
    try {
        await ResumeExecution();
    } catch (e) {
        console.error('Failed to resume execution:', e);
    }
}

// 実行停止
export async function stopExecution(): Promise<void> {
    try {
        await StopExecution();
    } catch (e) {
        console.error('Failed to stop execution:', e);
    }
}
