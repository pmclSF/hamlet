import { describe, it, expect } from 'vitest';

function withTempResource(callback) {
  const resource = { id: Date.now(), active: true, data: null };
  try {
    return callback(resource);
  } finally {
    resource.active = false;
    resource.data = null;
  }
}

describe('ContextManagerPattern', () => {
  it('should provide resource within callback scope', () => {
    const result = withTempResource((res) => {
      res.data = 'temporary';
      expect(res.active).toBe(true);
      return res.data;
    });
    expect(result).toBe('temporary');
  });

  it('should clean up after callback completes', () => {
    const ref = {};
    withTempResource((res) => {
      ref.resource = res;
      expect(res.active).toBe(true);
    });
    expect(ref.resource.active).toBe(false);
    expect(ref.resource.data).toBeNull();
  });
});
