import { describe, it, expect, beforeEach } from 'vitest';
import { register } from '../../src/auth/register.js';
import { createTestUser } from '../fixtures/db.js';
import { expectUser } from '../helpers/assertions.js';

describe('register', () => {
  describe('new user', () => {
    it('should create a user', async () => {
      const user = await register('new@test.com', 'password123');
      expectUser(user);
      expect(user.email).toBe('new@test.com');
    });

    it('should return the user id', async () => {
      const user = await register('another@test.com', 'password123');
      expect(user.id).toBeDefined();
    });
  });

  describe('duplicate user', () => {
    beforeEach(async () => {
      await createTestUser('existing@test.com');
    });

    it('should reject duplicate email', async () => {
      await expect(register('existing@test.com', 'pass')).rejects.toThrow('already exists');
    });
  });

  describe('validation', () => {
    it('should reject invalid email', async () => {
      await expect(register('bad', 'password123')).rejects.toThrow('Invalid email');
    });
  });
});
