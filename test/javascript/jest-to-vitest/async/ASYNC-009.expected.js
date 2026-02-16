import { describe, it, expect } from 'vitest';

describe('Slow operations', () => {
  it('completes a long-running upload', async () => {
    const result = await uploadLargeFile('data.csv');
    expect(result.status).toBe('success');
    expect(result.bytesUploaded).toBeGreaterThan(0);
  }, 15000);

  it('processes a batch job', async () => {
    const report = await runBatchJob({ items: 1000 });
    expect(report.processed).toBe(1000);
    expect(report.failed).toBe(0);
  }, 30000);
});
