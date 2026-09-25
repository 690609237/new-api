// Model output is untrusted. The shared Markdown renderer sanitizes HTML;
// only the third (risk) column receives our controlled highlight markup.
export function highlightDailyReviewRisks(content: string): string {
  return content
    .split('\n')
    .map((line) => {
      if (!line.startsWith('|')) return line
      const cells = line.split('|')
      if (cells.length < 7) return line
      const risk = cells[3].trim().replaceAll(/<\/?mark>/g, '')
      if (risk !== '极高' && risk !== '高') return line
      cells[3] = ` <mark>${risk}</mark> `
      return cells.join('|')
    })
    .join('\n')
}
