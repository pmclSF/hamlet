import { describe, it, expect } from 'vitest';

describe('Buffer and typed array comparisons', () => {
  it('should compare equal buffers with toEqual', () => {
    const buf1 = Buffer.from('hello');
    const buf2 = Buffer.from('hello');
    expect(buf1).toEqual(buf2);
  });

  it('should detect different buffers', () => {
    const buf1 = Buffer.from('hello');
    const buf2 = Buffer.from('world');
    expect(buf1).not.toEqual(buf2);
  });

  it('should compare Uint8Array instances', () => {
    const arr1 = new Uint8Array([1, 2, 3, 4]);
    const arr2 = new Uint8Array([1, 2, 3, 4]);
    expect(arr1).toEqual(arr2);
  });

  it('should compare Float64Array instances', () => {
    const arr1 = new Float64Array([1.1, 2.2, 3.3]);
    const arr2 = new Float64Array([1.1, 2.2, 3.3]);
    expect(arr1).toEqual(arr2);
  });

  it('should check buffer length', () => {
    const buf = Buffer.from('test');
    expect(buf).toHaveLength(4);
  });
});
