
import { writable } from 'svelte/store';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { Logger } from '../services/logger';

const log = Logger.withComponent('LogStore');

export interface LogEntry {
    id: string; // Unique ID only for UI keys
    taskId: string;
    stream: 'stdout' | 'stderr';
    line: string;
    timestamp: string;
}

const MAX_LOGS = 1000;

function createLogStore() {
    const { subscribe, update, set } = writable<LogEntry[]>([]);

    return {
        subscribe,
        
        addLog: (entry: Omit<LogEntry, 'id'>) => {
            update((logs) => {
                const newEntry = { ...entry, id: crypto.randomUUID() };
                const newLogs = [...logs, newEntry];
                if (newLogs.length > MAX_LOGS) {
                    return newLogs.slice(newLogs.length - MAX_LOGS);
                }
                return newLogs;
            });
        },

        clear: () => set([]),
    };
}

export const logs = createLogStore();

// Derived store generator for specific task
import { derived } from 'svelte/store';
export const getTaskLogs = (taskId: string) => derived(logs, $logs => $logs.filter(l => l.taskId === taskId));

interface TaskLogEvent {
    taskId: string;
    stream: 'stdout' | 'stderr';
    line: string;
    timestamp: string;
}

export function initLogEvents() {
    EventsOn('task:log', (event: TaskLogEvent) => {
        logs.addLog({
            taskId: event.taskId,
            stream: event.stream,
            line: event.line,
            timestamp: event.timestamp || new Date().toISOString(),
        });
    });
    
    log.info('Log events initialized');
}
