// HIGH FANOUT FIXTURE: used by API and integration tests
import { setupRoutes } from '../../src/api/routes.js';
import { authMiddleware, rateLimiter } from '../../src/api/middleware.js';
import { getConfig } from '../../src/config/app.js';

export function createTestApp() {
  const routes: any[] = [];
  const app = {
    post: (path: string, handler: any) => routes.push({ method: 'POST', path, handler }),
    get: (path: string, handler: any) => routes.push({ method: 'GET', path, handler }),
    use: () => {},
  };
  setupRoutes(app);
  return { app, routes };
}

export function createTestMiddleware() {
  return {
    auth: authMiddleware(),
    limiter: rateLimiter(),
  };
}

export function getTestConfig() {
  return getConfig();
}
