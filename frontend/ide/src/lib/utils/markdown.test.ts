import { describe, it, expect } from 'vitest';
import { stripMarkdown } from './markdown';

describe('stripMarkdown', () => {
  it('should remove single backticks', () => {
    expect(stripMarkdown('Implement `auth` module')).toBe('Implement auth module');
  });

  it('should remove triple backticks', () => {
    expect(stripMarkdown('```Fix login bug```')).toBe('Fix login bug');
  });

  it('should handle empty strings', () => {
    expect(stripMarkdown('')).toBe('');
  });

  it('should handle null/undefined input', () => {
    expect(stripMarkdown(null as any)).toBe('');
    expect(stripMarkdown(undefined as any)).toBe('');
  });

  it('should preserve text without markdown', () => {
    expect(stripMarkdown('Normal task title')).toBe('Normal task title');
  });

  it('should handle mixed backticks', () => {
    expect(stripMarkdown('Fix `bug` in ```module```')).toBe('Fix bug in module');
  });
});
