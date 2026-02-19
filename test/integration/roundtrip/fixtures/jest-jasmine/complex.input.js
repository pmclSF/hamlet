import { describe, it, expect, beforeEach, afterEach, beforeAll, afterAll, jest } from '@jest/globals';
import { AuthService } from '../../src/auth-service.js';
import { TokenStore } from '../../src/token-store.js';
import { HttpClient } from '../../src/http-client.js';

// Tests for AuthService including token refresh and session management
describe('AuthService', () => {
  let authService;
  let httpSpy;

  beforeAll(() => {
    process.env.AUTH_ENDPOINT = 'https://auth.test.local';
  });

  afterAll(() => {
    delete process.env.AUTH_ENDPOINT;
  });

  beforeEach(() => {
    authService = new AuthService(new TokenStore());
    httpSpy = jest.spyOn(HttpClient.prototype, 'post');
  });

  afterEach(() => {
    httpSpy.mockRestore();
  });

  it('should authenticate with valid credentials', async () => {
    httpSpy.mockResolvedValue({ token: 'abc-123', expiresIn: 3600 });
    const session = await authService.login('admin', 'secret');
    expect(session.token).toBe('abc-123');
    expect(session.isAuthenticated).toBe(true);
  });

  it('should reject invalid credentials', async () => {
    httpSpy.mockRejectedValue(new Error('401 Unauthorized'));
    await expect(authService.login('admin', 'wrong')).rejects.toThrow('401 Unauthorized');
  });

  it('should report unauthenticated before login', () => {
    expect(authService.isAuthenticated()).toBe(false);
  });

  describe('token management', () => {
    beforeEach(async () => {
      httpSpy.mockResolvedValue({ token: 'abc-123', expiresIn: 3600 });
      await authService.login('admin', 'secret');
    });

    it('should store the token after login', () => {
      expect(authService.getToken()).toBe('abc-123');
    });

    it('should refresh an expired token', async () => {
      httpSpy.mockResolvedValue({ token: 'def-456', expiresIn: 3600 });
      const newSession = await authService.refreshToken();
      expect(newSession.token).toBe('def-456');
    });

    it('should clear the token on logout', () => {
      authService.logout();
      expect(authService.getToken()).toBeNull();
    });
  });

  describe('session persistence', () => {
    it('should restore a session from stored token', async () => {
      httpSpy.mockResolvedValue({ valid: true, userId: 7 });
      const restored = await authService.restoreSession('saved-token-xyz');
      expect(restored.userId).toBe(7);
      expect(authService.isAuthenticated()).toBe(true);
    });

    // Test using done callback pattern for async completion
    it('should emit a session-expired event on timeout', (done) => {
      authService.onSessionExpired(() => {
        expect(authService.isAuthenticated()).toBe(false);
        done();
      });
      authService.simulateTimeout();
    });
  });
});
