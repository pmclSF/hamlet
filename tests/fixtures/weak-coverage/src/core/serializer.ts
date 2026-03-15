export function toJSON(data: unknown): string {
  return JSON.stringify(data, null, 2);
}

export function toCSV(rows: string[][]): string {
  return rows.map(row => row.join(',')).join('\n');
}

export function toXML(tag: string, content: string): string {
  return `<${tag}>${content}</${tag}>`;
}
