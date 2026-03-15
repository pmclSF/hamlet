import { describe, it, expect, beforeEach } from 'vitest';
import { register } from '../../src/auth/register.js';
import { login } from '../../src/auth/login.js';
import { createSession, getSession } from '../../src/auth/session.js';
import { seedDatabase, cleanDatabase } from '../fixtures/db.js';
import { setupAuthenticatedUser } from '../fixtures/auth.js';
import { setupTestEnvironment } from '../helpers/setup.js';
import { expectUser, expectSession } from '../helpers/assertions.js';

describe('auth flow', () => {
  beforeEach(async () => {
    const env = await setupTestEnvironment();
  });

  describe('register then login', () => {
    it('should register and then login', async () => {
      const registered = await register('flow@test.com', 'password123');
      expectUser(registered);

      const loggedIn = await login('flow@test.com', 'password123');
      expectUser(loggedIn);
      expect(loggedIn.email).toBe('flow@test.com');
    });

    it('should create session after login', async () => {
      await register('session-flow@test.com', 'password123');
      const user = await login('session-flow@test.com', 'password123');
      const token = await createSession(user.id);
      expectSession(token);

      const sessionUser = await getSession(token);
      expect(sessionUser).toBe(user.id);
    });
  });

  describe('fixture-based auth', () => {
    it('should use pre-configured authenticated user', async () => {
      const { user, token } = await setupAuthenticatedUser();
      expectUser(user);
      expectSession(token);
    });
  });
});
