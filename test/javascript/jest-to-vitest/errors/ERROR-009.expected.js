import { describe, it, expect } from 'vitest';

describe('StreamProcessor', () => {
  it('should validate each chunk in the callback', (done) => {
    const stream = createStream([1, 2, 3]);
    const results = [];
    stream.on('data', (chunk) => {
      expect(chunk).toBeGreaterThan(0);
      results.push(chunk);
    });
    stream.on('end', () => {
      expect(results).toHaveLength(3);
      done();
    });
  });

  it('should assert inside a forEach callback', () => {
    const items = getProcessedItems();
    items.forEach((item) => {
      expect(item).toHaveProperty('id');
      expect(item.id).toBeGreaterThan(0);
    });
  });

  it('should validate inside a map transform', () => {
    const raw = [{ val: '10' }, { val: '20' }, { val: '30' }];
    const parsed = raw.map((item) => {
      const num = parseInt(item.val, 10);
      expect(num).not.toBeNaN();
      return num;
    });
    expect(parsed).toEqual([10, 20, 30]);
  });
});
