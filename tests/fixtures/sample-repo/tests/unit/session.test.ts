import { describe, it, expect } from 'vitest';
import { createSession, getSession } from '../../src/auth/session.js';
import { setupAuthenticatedUser } from '../fixtures/auth.js';
import { expectSession } from '../helpers/assertions.js';

describe('session', () => {
  describe('createSession', () => {
    it('should return a session token', async () => {
      const token = await createSession('user_1');
      expectSession(token);
    });

    it('should create unique tokens', async () => {
      const token1 = await createSession('user_1');
      const token2 = await createSession('user_1');
      expect(token1).not.toBe(token2);
    });
  });

  describe('getSession', () => {
    it('should retrieve session by token', async () => {
      const token = await createSession('user_42');
      const userId = await getSession(token);
      expect(userId).toBe('user_42');
    });

    it('should return null for unknown token', async () => {
      const result = await getSession('invalid_token');
      expect(result).toBeNull();
    });
  });

  describe('authenticated user flow', () => {
    it('should create session from fixture user', async () => {
      const { token } = await setupAuthenticatedUser();
      expectSession(token);
    });
  });
});
