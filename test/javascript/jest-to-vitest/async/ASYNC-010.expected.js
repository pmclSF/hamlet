import { describe, it, expect } from 'vitest';

describe('Async iterators', () => {
  it('consumes an async iterator', async () => {
    const results = [];
    for await (const item of fetchPaginatedResults('/api/items')) {
      results.push(item);
    }
    expect(results).toHaveLength(10);
    expect(results[0]).toHaveProperty('id');
  });

  it('handles early termination of async iterator', async () => {
    const results = [];
    for await (const item of fetchPaginatedResults('/api/items')) {
      results.push(item);
      if (results.length >= 3) break;
    }
    expect(results).toHaveLength(3);
  });
});
