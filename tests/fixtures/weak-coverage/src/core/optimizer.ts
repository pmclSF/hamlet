export function deduplicate(items: string[]): string[] {
  return [...new Set(items)];
}

export function sortByLength(items: string[]): string[] {
  return [...items].sort((a, b) => a.length - b.length);
}

export function filterEmpty(items: string[]): string[] {
  return items.filter(item => item.trim().length > 0);
}
