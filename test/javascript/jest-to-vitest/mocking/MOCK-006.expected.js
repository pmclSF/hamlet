import { describe, it, expect, vi } from 'vitest';

vi.mock('./utils', () => ({
  ...await vi.importActual('./utils'),
  formatDate: vi.fn(() => '2024-01-01'),
  generateId: vi.fn(() => 'mock-id-123'),
}));

const { formatDate, generateId, parseInput } = require('./utils');

describe('Partial module mock', () => {
  it('uses mocked formatDate', () => {
    expect(formatDate(new Date())).toBe('2024-01-01');
  });

  it('uses mocked generateId', () => {
    expect(generateId()).toBe('mock-id-123');
  });

  it('uses real parseInput', () => {
    expect(parseInput('hello')).toBeDefined();
  });
});
