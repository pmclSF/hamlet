export function toUpperCase(input: string): string {
  return input.toUpperCase();
}

export function toCamelCase(input: string): string {
  return input.replace(/-([a-z])/g, (_, c) => c.toUpperCase());
}

export function toSnakeCase(input: string): string {
  return input.replace(/[A-Z]/g, c => '_' + c.toLowerCase());
}

export function truncate(input: string, maxLength: number): string {
  return input.length > maxLength ? input.substring(0, maxLength) + '...' : input;
}
