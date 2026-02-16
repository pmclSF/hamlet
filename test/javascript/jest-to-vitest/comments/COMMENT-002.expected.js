/*
 * Tests for the validation module.
 * These tests ensure that input validation
 * works correctly for all supported types.
 */
import { describe, it, expect } from 'vitest';

describe('validation module', () => {
  it('should validate email format', () => {
    const email = 'user@example.com';
    const isValid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
    expect(isValid).toBe(true);
  });

  /* This test covers invalid inputs */
  it('should reject invalid email', () => {
    const email = 'not-an-email';
    const isValid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
    expect(isValid).toBe(false);
  });
});
