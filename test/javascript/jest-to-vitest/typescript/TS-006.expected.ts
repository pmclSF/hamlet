import { describe, it, expect, vi } from 'vitest';

interface User {
  id: number;
  name: string;
  email: string;
}

interface ApiResponse {
  data: User[];
  total: number;
}

describe('API', () => {
  it('should mock typed response', () => {
    const mockFetch = vi.fn<() => Promise<ApiResponse>>().mockResolvedValue({
      data: [{ id: 1, name: 'Alice', email: 'a@b.com' }],
      total: 1,
    });
    expect(mockFetch).toBeDefined();
  });

  it('should mock simple return type', () => {
    const mockGetName = vi.fn<() => string>().mockReturnValue('Alice');
    expect(mockGetName()).toBe('Alice');
  });
});
