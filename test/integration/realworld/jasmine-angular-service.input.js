// Jasmine test for an Angular authentication service
// Inspired by real-world Angular service tests with dependency injection

import { AuthService } from './auth.service.js';
import { TokenStorage } from '../storage/token-storage.js';
import { HttpClient } from '../http/http-client.js';

describe('AuthService', () => {
  let authService;
  let httpSpy;
  let storageSpy;

  beforeEach(() => {
    httpSpy = jasmine.createSpyObj('HttpClient', ['post', 'get', 'setHeader']);
    storageSpy = jasmine.createSpyObj('TokenStorage', ['get', 'set', 'clear']);
    authService = new AuthService(httpSpy, storageSpy);
  });

  describe('login', () => {
    it('should send credentials to the login endpoint', async () => {
      httpSpy.post.and.returnValue(Promise.resolve({
        token: 'abc123',
        refreshToken: 'refresh456',
        expiresIn: 3600,
      }));

      await authService.login('admin@example.com', 'password123');

      expect(httpSpy.post).toHaveBeenCalledWith('/api/auth/login', {
        email: 'admin@example.com',
        password: 'password123',
      });
    });

    it('should store the access token on successful login', async () => {
      httpSpy.post.and.returnValue(Promise.resolve({ token: 'abc123' }));

      await authService.login('user@test.com', 'secret');

      expect(storageSpy.set).toHaveBeenCalledWith('accessToken', 'abc123');
    });

    it('should set the authorization header after login', async () => {
      httpSpy.post.and.returnValue(Promise.resolve({ token: 'xyz789' }));

      await authService.login('user@test.com', 'pass');

      expect(httpSpy.setHeader).toHaveBeenCalledWith('Authorization', 'Bearer xyz789');
    });

    it('should reject with an error when credentials are invalid', async () => {
      httpSpy.post.and.returnValue(Promise.reject(new Error('401 Unauthorized')));

      try {
        await authService.login('wrong@test.com', 'bad');
        fail('Expected login to throw');
      } catch (err) {
        expect(err.message).toBe('401 Unauthorized');
      }
    });
  });

  describe('logout', () => {
    it('should clear stored tokens on logout', () => {
      authService.logout();

      expect(storageSpy.clear).toHaveBeenCalled();
    });
  });

  describe('isAuthenticated', () => {
    it('should return true when a valid token exists', () => {
      storageSpy.get.and.returnValue('valid-token');

      expect(authService.isAuthenticated()).toBe(true);
    });

    it('should return false when no token is stored', () => {
      storageSpy.get.and.returnValue(null);

      expect(authService.isAuthenticated()).toBe(false);
    });
  });
});
