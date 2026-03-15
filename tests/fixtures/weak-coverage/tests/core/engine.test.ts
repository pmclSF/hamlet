import { describe, it, expect } from 'vitest';
import { initialize, process, validate } from '../../src/core/engine';

describe('Engine', () => {
  it('should initialize with config', () => {
    const result = initialize({ key: 'value' });
    expect(result.ready).toBe(true);
  });

  it('should not initialize with empty config', () => {
    const result = initialize({});
    expect(result.ready).toBe(false);
  });

  it('should process input to lowercase', () => {
    expect(process('  HELLO  ')).toBe('hello');
  });

  it('should validate non-empty input', () => {
    expect(validate('hello')).toBe(true);
    expect(validate('')).toBe(false);
  });
});
