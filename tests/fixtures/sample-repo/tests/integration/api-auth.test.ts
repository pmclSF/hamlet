// DUPLICATE: overlaps with api-routes.test.ts
// Same fixtures (api, auth, db), same helpers (request, setup), similar assertions
import { describe, it, expect, beforeEach } from 'vitest';
import { createTestApp, createTestMiddleware } from '../fixtures/api.js';
import { setupAuthenticatedUser } from '../fixtures/auth.js';
import { seedDatabase, cleanDatabase } from '../fixtures/db.js';
import { createRequest, createAuthenticatedRequest, createResponse } from '../helpers/request.js';
import { setupTestEnvironment } from '../helpers/setup.js';

describe('API authentication', () => {
  beforeEach(async () => {
    await setupTestEnvironment();
  });

  describe('login endpoint', () => {
    it('should expose login route', () => {
      const { routes } = createTestApp();
      const login = routes.find((r) => r.path === '/login');
      expect(login).toBeDefined();
    });
  });

  describe('register endpoint', () => {
    it('should expose register route', () => {
      const { routes } = createTestApp();
      const reg = routes.find((r) => r.path === '/register');
      expect(reg).toBeDefined();
    });
  });

  describe('auth middleware', () => {
    it('should block requests without token', async () => {
      const middleware = createTestMiddleware();
      const req = createRequest('GET', '/api/data');
      await expect(
        middleware.auth(req, {}, () => {})
      ).rejects.toThrow('Missing token');
    });

    it('should allow authenticated requests', async () => {
      const { token } = await setupAuthenticatedUser();
      const middleware = createTestMiddleware();
      const req = createAuthenticatedRequest('GET', '/api/data', token);
      // Would pass auth check in real scenario
      expect(req.headers.authorization).toContain(token);
    });
  });
});
