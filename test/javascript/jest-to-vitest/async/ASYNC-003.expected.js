import { describe, it, expect } from 'vitest';

describe('UserAPI', () => {
  it('fetches user with callback', (done) => {
    fetchUser(1, (err, data) => {
      expect(err).toBeNull();
      expect(data).toBeDefined();
      expect(data.name).toBe('Alice');
      done();
    });
  });

  it('handles callback errors', (done) => {
    fetchUser(-1, (err) => {
      expect(err).toBeDefined();
      expect(err.message).toBe('User not found');
      done();
    });
  });
});
