import { writable } from 'svelte/store';

export type WindowId = 'chat' | 'process' | 'wbs' | 'backlog';

export interface WindowState {
  id: WindowId;
  title: string;
  isOpen: boolean;
  isMinimized: boolean;
  position: { x: number; y: number };
  size?: { width: number; height: number };
  zIndex: number;
}

const DEFAULT_WINDOWS: Record<WindowId, Omit<WindowState, 'zIndex'>> = {
  chat: {
    id: 'chat',
    title: 'Chat',
    isOpen: true,
    isMinimized: false,
    position: { x: window.innerWidth - 620, y: window.innerHeight - 500 }, // Default bottom right approx
    size: { width: 600, height: 450 }
  },
  process: {
    id: 'process',
    title: 'Process & Resources',
    isOpen: false,
    isMinimized: false,
    position: { x: 20, y: 20 },
    size: { width: 400, height: 500 }
  },
  wbs: {
    id: 'wbs',
    title: 'Work Breakdown Structure',
    isOpen: false,
    isMinimized: false,
    position: { x: 300, y: 100 },
    size: { width: 800, height: 600 }
  },
  backlog: {
    id: 'backlog',
    title: 'Backlog',
    isOpen: false,
    isMinimized: false,
    position: { x: 50, y: 100 },
    size: { width: 350, height: 500 }
  }
};

function createWindowStore() {
  const { subscribe, update } = writable<Record<WindowId, WindowState>>({
    chat: { ...DEFAULT_WINDOWS.chat, zIndex: 100 },
    process: { ...DEFAULT_WINDOWS.process, zIndex: 101 },
    wbs: { ...DEFAULT_WINDOWS.wbs, zIndex: 102 },
    backlog: { ...DEFAULT_WINDOWS.backlog, zIndex: 103 },
  });

  let maxZIndex = 110;

  return {
    subscribe,
    open: (id: WindowId) => update(windows => {
      const window = windows[id];
      if (!window) return windows;
      return {
        ...windows,
        [id]: {
          ...window,
          isOpen: true,
          isMinimized: false,
          zIndex: ++maxZIndex
        }
      };
    }),
    close: (id: WindowId) => update(windows => {
      const window = windows[id];
      if (!window) return windows;
      return {
        ...windows,
        [id]: { ...window, isOpen: false }
      };
    }),
    toggle: (id: WindowId) => update(windows => {
      const window = windows[id];
      if (!window) return windows;
      
      // If closed, open and bring to front
      if (!window.isOpen) {
        return {
          ...windows,
          [id]: {
            ...window,
            isOpen: true,
            isMinimized: false,
            zIndex: ++maxZIndex
          }
        };
      }
      
      // If open but minimized, unminimize and bring to front
      if (window.isMinimized) {
        return {
          ...windows,
          [id]: {
             ...window,
             isMinimized: false,
             zIndex: ++maxZIndex
          }
        };
      }
      
      // If open and active, minimize? Or close? User habit usually toggle = open/close. 
      // Let's make it toggle open/close for taskbar behavior.
      return {
        ...windows,
        [id]: {
          ...window,
          isOpen: false
        }
      };
    }),
    minimize: (id: WindowId, minimized: boolean) => update(windows => {
       const window = windows[id];
       if (!window) return windows;
       return {
         ...windows,
         [id]: { ...window, isMinimized: minimized }
       };
    }),
    bringToFront: (id: WindowId) => update(windows => {
      const window = windows[id];
      if (!window || !window.isOpen) return windows;
      if (window.zIndex === maxZIndex) return windows;
      
      return {
        ...windows,
        [id]: { ...window, zIndex: ++maxZIndex }
      };
    }),
    updatePosition: (id: WindowId, x: number, y: number) => update(windows => {
      const window = windows[id];
      if (!window) return windows;
      return {
        ...windows,
        [id]: { ...window, position: { x, y } }
      };
    }),
    updateSize: (id: WindowId, width: number, height: number) => update(windows => {
      const window = windows[id];
      if (!window) return windows;
      return {
        ...windows,
        [id]: { ...window, size: { width, height } }
      };
    }),
    // Expose update for internal/storybook usage
    update
  };
}

export const windowStore = createWindowStore();
