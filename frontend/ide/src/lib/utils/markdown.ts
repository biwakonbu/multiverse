
/**
 * Strips markdown syntax from the input string to make it safe for direct display.
 * Primary focus is removing code block delimiters which are distracting in titles.
 */
export function stripMarkdown(text: string): string {
  if (!text) return "";
  
  // Remove triple backticks
  let clean = text.replace(/```/g, '');
  
  // Remove single backticks
  clean = clean.replace(/`/g, '');
  
  // Potential improvement: handle other markdown like **bold** or *italic* if needed,
  // but for now the user specifically complained about code blocks (` ``` `).
  
  return clean;
}
