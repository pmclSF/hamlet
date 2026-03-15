import { describe, it, expect, beforeEach } from 'vitest';
import { createTestApp, createTestMiddleware } from '../fixtures/api.js';
import { setupAuthenticatedUser } from '../fixtures/auth.js';
import { seedDatabase, cleanDatabase } from '../fixtures/db.js';
import { createRequest, createAuthenticatedRequest, createResponse } from '../helpers/request.js';
import { setupTestEnvironment } from '../helpers/setup.js';

describe('API routes', () => {
  beforeEach(async () => {
    await setupTestEnvironment();
  });

  describe('POST /login', () => {
    it('should have login route', () => {
      const { routes } = createTestApp();
      const loginRoute = routes.find((r) => r.path === '/login');
      expect(loginRoute).toBeDefined();
      expect(loginRoute?.method).toBe('POST');
    });
  });

  describe('POST /register', () => {
    it('should have register route', () => {
      const { routes } = createTestApp();
      const registerRoute = routes.find((r) => r.path === '/register');
      expect(registerRoute).toBeDefined();
      expect(registerRoute?.method).toBe('POST');
    });
  });

  describe('middleware', () => {
    it('should create auth middleware', () => {
      const middleware = createTestMiddleware();
      expect(middleware.auth).toBeDefined();
      expect(typeof middleware.auth).toBe('function');
    });

    it('should create rate limiter', () => {
      const middleware = createTestMiddleware();
      expect(middleware.limiter).toBeDefined();
      expect(typeof middleware.limiter).toBe('function');
    });

    it('should reject unauthenticated requests', async () => {
      const middleware = createTestMiddleware();
      const req = createRequest('GET', '/protected');
      await expect(
        middleware.auth(req, {}, () => {})
      ).rejects.toThrow('Missing token');
    });
  });
});
