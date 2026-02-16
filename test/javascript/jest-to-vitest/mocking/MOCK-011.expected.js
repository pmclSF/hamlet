import { describe, it, expect, vi, afterEach } from 'vitest';

describe('Date mocking', () => {
  it('mocks Date.now', () => {
    vi.spyOn(Date, 'now').mockReturnValue(1234567890000);
    const timestamp = Date.now();
    expect(timestamp).toBe(1234567890000);
  });

  it('mocks the Date constructor', () => {
    const fixedDate = new Date('2024-01-15T00:00:00Z');
    vi.spyOn(global, 'Date').mockImplementation(() => fixedDate);
    const now = new Date();
    expect(now.toISOString()).toBe('2024-01-15T00:00:00.000Z');
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });
});
