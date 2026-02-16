import { describe, it, expect } from 'vitest';

describe('FeatureX', () => {
  it('should work under normal conditions', () => {
    const result = 2 * 3;
    expect(result).toBe(6);
  });

  it.skip('should handle the deprecated API', () => {
    const legacy = { version: 1, deprecated: true };
    expect(legacy.deprecated).toBe(true);
    expect(legacy.version).toBeLessThan(2);
  });

  xit('should support XML input format', () => {
    const xml = '<root><item>test</item></root>';
    expect(xml).toContain('<item>');
  });

  it('should process JSON input', () => {
    const json = JSON.parse('{"key": "value"}');
    expect(json.key).toBe('value');
  });
});
