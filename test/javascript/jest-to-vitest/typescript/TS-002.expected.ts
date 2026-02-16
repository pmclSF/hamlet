import { describe, it, expect } from 'vitest';

describe('Parser', () => {
  it('should parse response', () => {
    const raw = JSON.parse('{"status": 200}');
    const response = raw as { status: number };
    expect(response.status).toBe(200);
  });

  it('should parse array response', () => {
    const raw = JSON.parse('[1, 2, 3]');
    const items = raw as number[];
    expect(items.length).toBe(3);
  });
});
