import { describe, it, expect, vi } from 'vitest';

vi.mock('./utils', () => {
  const actual = await vi.importActual('./utils');
  return {
    ...actual,
    formatDate: vi.fn(() => '2024-01-01'),
  };
});

describe('Utils', () => {
  it('uses real helpers but mocked formatDate', () => {
    expect(formatDate()).toBe('2024-01-01');
    expect(parseInput('test')).toBeDefined();
  });
});
