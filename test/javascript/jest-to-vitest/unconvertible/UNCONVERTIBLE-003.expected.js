// HAMLET-WARNING: Vitest uses `retry` in config or per-describe options instead of jest.retryTimes(3). Configure retry in vitest.config.ts or use describe("suite", () => {}, { retry: 3 }).
// Original: jest.retryTimes(3);
// jest.retryTimes(3);

import { describe, it, expect } from 'vitest';

describe('Flaky test', () => {
  it('sometimes fails due to timing', () => {
    const value = Math.random();
    expect(value).toBeLessThan(0.9);
  });

  it('sometimes fails due to external dependency', async () => {
    const response = await fetchUnstableEndpoint();
    expect(response.status).toBe(200);
  });
});
