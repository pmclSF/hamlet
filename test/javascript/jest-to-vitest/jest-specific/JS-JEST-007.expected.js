import { describe, it, expect, vi } from 'vitest';

vi.setConfig({ testTimeout: 60000 });

describe('Integration tests', () => {
  it('connects to the external API', async () => {
    const response = await fetchFromExternalAPI('/health');
    expect(response.status).toBe(200);
  });

  it('processes large payloads', async () => {
    const data = generateLargePayload(10000);
    const result = await processPayload(data);
    expect(result.success).toBe(true);
  });
});
