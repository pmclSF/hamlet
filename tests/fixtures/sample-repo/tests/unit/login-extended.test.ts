// DUPLICATE: overlaps heavily with login.test.ts
// Same fixtures (db), same helpers (assertions), similar suite path, similar assertions
import { describe, it, expect, beforeEach } from 'vitest';
import { login } from '../../src/auth/login.js';
import { createTestUser } from '../fixtures/db.js';
import { expectUser } from '../helpers/assertions.js';

describe('login', () => {
  beforeEach(async () => {
    await createTestUser('extended@test.com');
  });

  describe('successful login', () => {
    it('should authenticate valid user', async () => {
      const user = await login('extended@test.com', 'testpass');
      expectUser(user);
      expect(user.email).toBe('extended@test.com');
    });

    it('should return user with id field', async () => {
      const user = await login('extended@test.com', 'testpass');
      expect(user.id).toBeDefined();
      expect(typeof user.id).toBe('string');
    });
  });

  describe('failed login', () => {
    it('should fail with bad email format', async () => {
      await expect(login('invalid', 'testpass')).rejects.toThrow('Invalid email');
    });

    it('should fail with wrong password', async () => {
      await expect(login('extended@test.com', 'badpass')).rejects.toThrow('Invalid password');
    });
  });
});
