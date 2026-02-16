import { describe, it, expect } from 'vitest';

describe.concurrent('ParallelDataProcessing', () => {
  it('should process batch A', async () => {
    const batch = Array.from({ length: 100 }, (_, i) => i * 2);
    const sum = batch.reduce((a, b) => a + b, 0);
    expect(sum).toBe(9900);
  });

  it('should process batch B', async () => {
    const batch = Array.from({ length: 100 }, (_, i) => i * 3);
    const sum = batch.reduce((a, b) => a + b, 0);
    expect(sum).toBe(14850);
  });

  it('should process batch C', async () => {
    const batch = Array.from({ length: 50 }, (_, i) => i + 1);
    const product = batch.reduce((a, b) => a + b, 0);
    expect(product).toBe(1275);
  });

  it('should merge all batch results', async () => {
    const results = [9900, 14850, 1275];
    const total = results.reduce((a, b) => a + b, 0);
    expect(total).toBe(26025);
  });
});
