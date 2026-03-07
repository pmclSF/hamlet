import { formatDate } from '../utils/date';
import { describe, it, expect } from 'vitest';

describe('formatDate', () => {
  it('formats ISO dates', () => {
    expect(formatDate('2026-01-01')).toBe('Jan 1, 2026');
  });
});
