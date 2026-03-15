// DUPLICATE: overlaps heavily with register.test.ts
// Same fixtures (db), same helpers (assertions), near-identical suite path
import { describe, it, expect, beforeEach } from 'vitest';
import { register } from '../../src/auth/register.js';
import { createTestUser } from '../fixtures/db.js';
import { expectUser } from '../helpers/assertions.js';

describe('register', () => {
  describe('creating new user', () => {
    it('should successfully register', async () => {
      const user = await register('fresh@test.com', 'password123');
      expectUser(user);
      expect(user.email).toBe('fresh@test.com');
    });

    it('should assign an id', async () => {
      const user = await register('withid@test.com', 'password123');
      expect(user.id).toBeDefined();
    });
  });

  describe('existing user', () => {
    beforeEach(async () => {
      await createTestUser('taken@test.com');
    });

    it('should reject taken email', async () => {
      await expect(register('taken@test.com', 'pass')).rejects.toThrow('already exists');
    });
  });

  describe('input validation', () => {
    it('should reject malformed email', async () => {
      await expect(register('nope', 'password123')).rejects.toThrow('Invalid email');
    });
  });
});
