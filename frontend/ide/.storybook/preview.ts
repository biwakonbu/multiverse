import type { Preview } from '@storybook/svelte-vite';

// デザインシステムのCSS変数をインポート
import '../src/design-system/variables.css';

const preview: Preview = {
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i
      }
    },
    backgrounds: {
      options: {
        "multiverse-app":
        { name: 'multiverse-app', value: '#16181e' },
        dark: { name: 'dark', value: '#1a1a1a' },
        light: { name: 'light', value: '#ffffff' }
      }
    },
    // Docs ページのスタイル設定
    docs: {
      story: {
        inline: true
      }
    }
  },

  initialGlobals: {
    backgrounds: {
      value: 'multiverse-app'
    }
  }
};

export default preview;
