import { describe, it, expect } from 'vitest';

describe('SchemaValidator', () => {
  it('should throw TypeError when value is not a string', () => {
    expect(() => validateField('name', 42)).toThrow(TypeError);
  });

  it('should throw RangeError for negative quantities', () => {
    expect(() => validateField('quantity', -5)).toThrow(RangeError);
  });

  it('should throw TypeError for null schema definitions', () => {
    expect(() => {
      const validator = new SchemaValidator(null);
      validator.validate({ name: 'test' });
    }).toThrow(TypeError);
  });
});
