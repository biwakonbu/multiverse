import { describe, it, expect } from 'vitest';
import { readFileSync, readdirSync, statSync } from 'fs';
import { join } from 'path';

/**
 * This test ensures that stray markdown backticks (```) do not appear
 * in Svelte component files outside of script/style blocks.
 * 
 * Issue history: There was a bug where ``` appeared in the template portion
 * of WBSGraphNode.svelte and Toolbar.svelte, causing literal backticks
 * to be displayed in the IDE's main view.
 */
describe('Source Code Integrity', () => {
  const srcDir = join(__dirname, '..');

  function getAllSvelteFiles(dir: string): string[] {
    const files: string[] = [];
    const entries = readdirSync(dir);
    
    for (const entry of entries) {
      const fullPath = join(dir, entry);
      const stat = statSync(fullPath);
      
      if (stat.isDirectory() && entry !== 'node_modules') {
        files.push(...getAllSvelteFiles(fullPath));
      } else if (entry.endsWith('.svelte')) {
        files.push(fullPath);
      }
    }
    
    return files;
  }

  it('should not have stray backticks outside of script/style blocks in Svelte files', () => {
    const svelteFiles = getAllSvelteFiles(srcDir);
    const issues: { file: string; line: number; content: string }[] = [];

    for (const file of svelteFiles) {
      const content = readFileSync(file, 'utf-8');
      const lines = content.split('\n');
      
      let inScript = false;
      let inStyle = false;
      
      for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        
        // Track script/style blocks
        if (line.includes('<script')) inScript = true;
        if (line.includes('</script>')) inScript = false;
        if (line.includes('<style')) inStyle = true;
        if (line.includes('</style>')) inStyle = false;
        
        // Check for stray backticks in template section
        if (!inScript && !inStyle) {
          // Line that is ONLY backticks (with optional whitespace)
          if (/^\s*```\s*$/.test(line)) {
            issues.push({
              file: file.replace(srcDir, ''),
              line: i + 1,
              content: line
            });
          }
        }
      }
    }

    if (issues.length > 0) {
      const message = issues.map(
        issue => `${issue.file}:${issue.line} contains stray backticks: "${issue.content}"`
      ).join('\n');
      expect.fail(`Found stray backticks in template sections:\n${message}`);
    }
  });
});
