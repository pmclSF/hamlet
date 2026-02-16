import { describe, it, expect, beforeEach } from 'vitest';

describe('ApiClient', () => {
  let client;
  let baseUrl;

  beforeEach(async () => {
    baseUrl = 'https://api.example.com';
    const config = await Promise.resolve({
      timeout: 5000,
      retries: 3,
    });
    client = {
      baseUrl,
      timeout: config.timeout,
      retries: config.retries,
      async get(path) {
        return { status: 200, url: `${baseUrl}${path}` };
      },
    };
  });

  it('should be configured with the correct base URL', () => {
    expect(client.baseUrl).toBe('https://api.example.com');
  });

  it('should have the correct timeout', () => {
    expect(client.timeout).toBe(5000);
  });

  it('should construct full URLs from paths', async () => {
    const response = await client.get('/users');
    expect(response.url).toBe('https://api.example.com/users');
  });
});
