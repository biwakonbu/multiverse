import { describe, it, expect, afterEach } from 'vitest';
import { cleanup, render } from '@testing-library/svelte';
import Button from './components/Button.svelte';
import Badge from './components/Badge.svelte';
import Flex from './components/Flex.svelte';

describe('Design System', () => {
  afterEach(() => {
    cleanup();
  });

  describe('Button', () => {
    it('renders label correctly', () => {
      const { getByText } = render(Button, { label: 'Click Me' });
      expect(getByText('Click Me')).toBeTruthy();
    });

    it('renders crystal variant', () => {
      const { container } = render(Button, { variant: 'crystal', label: 'Gem' });
      const btn = container.querySelector('button');
      expect(btn?.classList.contains('variant-crystal')).toBe(true);
    });
  });

  describe('Badge', () => {
    it('renders status label', () => {
      const { getByText } = render(Badge, { status: 'running' });
      expect(getByText('RUNNING')).toBeTruthy();
    });

    it('applies correct class for status', () => {
      const { container } = render(Badge, { status: 'failed' });
      const span = container.querySelector('span');
      expect(span?.classList.contains('status-failed')).toBe(true);
    });
  });

  describe('Flex', () => {
    it('renders slots', () => {
      // Basic check to see if it renders content
      // Note: testing CSS variables is harder in JSDOM dependent on style engine, 
      // but we can check if attributes or classes are applied if we implemented them that way.
      // Svelte creates scoped classes.
      // We will check if it runs without error.
      const { container } = render(Flex);
      expect(container.querySelector('div')).toBeTruthy();
    });
  });
});
