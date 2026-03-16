import { describe, it, expect } from 'vitest';
import { createSession, destroySession, getSessionUser } from '../../../src/auth/session';

describe('createSession', () => {
  it('should create session with valid token', () => {
    const result = createSession('tok_test');
    expect(result.sessionId).toContain('sess_');
  });

  it('should throw for invalid token', () => {
    expect(() => createSession('invalid')).toThrow('Invalid token');
  });
});

describe('destroySession', () => {
  it('should destroy session', () => {
    expect(destroySession('sess_1').destroyed).toBe(true);
  });
});

describe('getSessionUser', () => {
  it('should return user for session', () => {
    expect(getSessionUser('sess_1').userId).toBe('user_1');
  });
});
