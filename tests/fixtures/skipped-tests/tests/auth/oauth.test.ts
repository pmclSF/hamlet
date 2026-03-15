import { describe, it, expect } from 'vitest';
import { initiateOAuth, exchangeCode, refreshAccessToken } from '../../src/auth/oauth';

describe('OAuth', () => {
  it('should generate authorization URL', () => {
    const url = initiateOAuth('google');
    expect(url).toContain('google.com');
  });

  it.skip('should exchange authorization code for tokens', () => {
    // TODO: requires OAuth provider sandbox
    const tokens = exchangeCode('auth_code_123');
    expect(tokens.accessToken).toBeDefined();
  });

  it.skip('should refresh expired access token', () => {
    // TODO: requires valid refresh token from provider
    const newToken = refreshAccessToken('refresh_abc');
    expect(newToken).toContain('renewed');
  });

  it('should handle empty provider name', () => {
    const url = initiateOAuth('');
    expect(url).toContain('authorize');
  });
});
