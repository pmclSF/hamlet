import { describe, it, expect, vi } from 'vitest';

vi.setConfig({ testTimeout: 30000 });

describe('Slow API tests', () => {
  it('fetches a large dataset', async () => {
    const data = await fetchLargeDataset();
    expect(data.length).toBeGreaterThan(1000);
  });

  it('performs batch processing', async () => {
    const results = await batchProcess(generateItems(500));
    expect(results.every((r) => r.processed)).toBe(true);
  });

  it('handles timeout gracefully', async () => {
    const result = await fetchWithTimeout('/slow-endpoint', 25000);
    expect(result).toBeDefined();
  });
});
