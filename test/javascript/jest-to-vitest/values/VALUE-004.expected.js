import { describe, it, expect } from 'vitest';

describe('regex pattern matching', () => {
  it('should match ISO date format', () => {
    const dateStr = '2024-01-15';
    expect(dateStr).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });

  it('should match email pattern', () => {
    const email = 'user@example.com';
    expect(email).toMatch(/^[^\s@]+@[^\s@]+\.[^\s@]+$/);
  });

  it('should match UUID format', () => {
    const uuid = '550e8400-e29b-41d4-a716-446655440000';
    expect(uuid).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/);
  });

  it('should not match invalid patterns', () => {
    const invalid = 'not-a-date';
    expect(invalid).not.toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });
});
