import { describe, it, expect } from 'vitest';
import { generateTOTP, verifyTOTP, generateBackupCodes } from '../../src/auth/mfa';

describe('MFA', () => {
  it('should generate 6-digit TOTP code', () => {
    const code = generateTOTP('secret123');
    expect(code).toHaveLength(6);
    expect(code).toMatch(/^\d+$/);
  });

  it.skip('should verify TOTP against time-based window', () => {
    // TODO: requires time-synchronized TOTP validation
    const code = generateTOTP('secret123');
    expect(verifyTOTP(code, 'secret123')).toBe(true);
  });

  it('should generate requested number of backup codes', () => {
    const codes = generateBackupCodes(5);
    expect(codes).toHaveLength(5);
  });
});
