import { describe, it, expect, beforeEach } from 'vitest';
import { login } from '../../src/auth/login.js';
import { createTestUser } from '../fixtures/db.js';
import { expectUser } from '../helpers/assertions.js';

describe('login', () => {
  beforeEach(async () => {
    await createTestUser('login@test.com');
  });

  describe('valid credentials', () => {
    it('should return a user object', async () => {
      const user = await login('login@test.com', 'testpass');
      expectUser(user);
      expect(user.email).toBe('login@test.com');
    });

    it('should include user id', async () => {
      const user = await login('login@test.com', 'testpass');
      expect(user.id).toBeDefined();
    });
  });

  describe('invalid credentials', () => {
    it('should reject invalid email', async () => {
      await expect(login('bad', 'testpass')).rejects.toThrow('Invalid email');
    });

    it('should reject wrong password', async () => {
      await expect(login('login@test.com', 'wrong')).rejects.toThrow('Invalid password');
    });

    it('should reject unknown user', async () => {
      await expect(login('nobody@test.com', 'pass')).rejects.toThrow('User not found');
    });
  });
});
