// Relies on globals set in jest.setup.js:
//   global.fetch = vi.fn()
//   global.API_BASE_URL = 'http://localhost:3000'

import { describe, it, expect, vi, beforeEach } from 'vitest';

describe('API client', () => {
  beforeEach(() => {
    fetch.mockClear();
    fetch.mockResolvedValue({
      ok: true,
      json: vi.fn().mockResolvedValue({ data: 'test' }),
    });
  });

  it('calls fetch with correct URL', async () => {
    await apiClient.get('/users');
    expect(fetch).toHaveBeenCalledWith('http://localhost:3000/users', expect.any(Object));
  });

  it('handles fetch errors', async () => {
    fetch.mockRejectedValue(new Error('Network error'));
    await expect(apiClient.get('/users')).rejects.toThrow('Network error');
  });
});
