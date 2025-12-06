/**
 * バックログ状態管理ストア
 *
 * バックログアイテムの取得・解決・削除を管理
 * Wails Events でリアルタイム更新
 */

import { writable, derived } from 'svelte/store';
import { EventsOn } from '../../wailsjs/wailsjs/runtime/runtime';
import {
    GetBacklogItems,
    ResolveBacklogItem as WailsResolveBacklogItem,
    DeleteBacklogItem as WailsDeleteBacklogItem,
    type BacklogItem,
} from '../../wailsjs/wailsjs/go/main/App';
import { Logger } from '../services/logger';

const log = Logger.withComponent('BacklogStore');

export type { BacklogItem };
export type BacklogType = 'FAILURE' | 'QUESTION' | 'BLOCKER';

// バックログアイテム一覧ストア
function createBacklogStore() {
    const { subscribe, set, update } = writable<BacklogItem[]>([]);

    return {
        subscribe,

        // アイテム一覧を設定
        setItems: (items: BacklogItem[]) => {
            log.info('backlog items updated', { count: items.length });
            set(items);
        },

        // アイテムを追加
        addItem: (item: BacklogItem) => {
            log.info('backlog item added', { id: item.id, type: item.type });
            update((items) => [...items, item]);
        },

        // アイテムを削除
        removeItem: (id: string) => {
            log.info('backlog item removed', { id });
            update((items) => items.filter((i) => i.id !== id));
        },

        // クリア
        clear: () => {
            log.info('backlog items cleared');
            set([]);
        },
    };
}

export const backlogItems = createBacklogStore();

// 未解決アイテム数
export const unresolvedCount = derived(backlogItems, ($items) => {
    return $items.filter((item) => !item.resolvedAt).length;
});

// タイプ別カウント
export const countsByType = derived(backlogItems, ($items) => {
    const counts: Record<BacklogType, number> = {
        FAILURE: 0,
        QUESTION: 0,
        BLOCKER: 0,
    };

    for (const item of $items) {
        if (!item.resolvedAt) {
            counts[item.type]++;
        }
    }

    return counts;
});

// バックログアイテムを読み込み
export async function loadBacklogItems(): Promise<void> {
    try {
        const items = await GetBacklogItems();
        backlogItems.setItems(items || []);
    } catch (error) {
        log.error('failed to load backlog items', { error });
    }
}

// バックログアイテムを解決
export async function resolveItem(id: string, resolution: string = 'Resolved'): Promise<void> {
    try {
        log.info('resolving backlog item', { id, resolution });
        await WailsResolveBacklogItem(id, resolution);
        // 成功したらリストから削除（または再読み込み）
        backlogItems.removeItem(id);
    } catch (error) {
        log.error('failed to resolve backlog item', { id, error });
        throw error;
    }
}

// バックログアイテムを削除
export async function deleteItem(id: string): Promise<void> {
    try {
        log.info('deleting backlog item', { id });
        await WailsDeleteBacklogItem(id);
        backlogItems.removeItem(id);
    } catch (error) {
        log.error('failed to delete backlog item', { id, error });
        throw error;
    }
}

// Wails イベントリスナー初期化
export function initBacklogEvents(): void {
    // backlog:added イベントをリッスン
    EventsOn('backlog:added', (item: BacklogItem) => {
        log.info('backlog item added via event', { id: item.id, type: item.type });
        backlogItems.addItem(item);
    });

    // 初期データを読み込み
    loadBacklogItems();

    log.info('backlog events initialized');
}
