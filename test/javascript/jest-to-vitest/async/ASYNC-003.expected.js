import { describe, it, expect } from 'vitest';

describe('UserAPI', () => {
// HAMLET-WARNING: done() callback pattern detected. Vitest supports done() but async/await is preferred. Consider refactoring to async/await.
// Original: it('fetches user with callback', (done) => {
  it('fetches user with callback', (done) => {
    fetchUser(1, (err, data) => {
      expect(err).toBeNull();
      expect(data).toBeDefined();
      expect(data.name).toBe('Alice');
      done();
    });
  });

// HAMLET-WARNING: done() callback pattern detected. Vitest supports done() but async/await is preferred. Consider refactoring to async/await.
// Original: it('handles callback errors', (done) => {
  it('handles callback errors', (done) => {
    fetchUser(-1, (err) => {
      expect(err).toBeDefined();
      expect(err.message).toBe('User not found');
      done();
    });
  });
});
