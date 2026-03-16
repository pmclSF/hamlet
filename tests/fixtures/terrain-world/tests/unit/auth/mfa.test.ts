import { describe, it, expect } from 'vitest';
import { generateMFAChallenge, verifyMFACode, enrollMFA } from '../../../src/auth/mfa';

describe('generateMFAChallenge', () => {
  it('should generate challenge', () => {
    const result = generateMFAChallenge('user_1');
    expect(result.challengeId).toContain('mfa_');
  });
});

describe('verifyMFACode', () => {
  it('should accept correct code', () => {
    expect(verifyMFACode('mfa_1', '123456')).toBe(true);
  });

  it('should reject wrong code', () => {
    expect(verifyMFACode('mfa_1', '000000')).toBe(false);
  });
});

describe('enrollMFA', () => {
  it('should enroll user', () => {
    expect(enrollMFA('user_1', 'totp').enrolled).toBe(true);
  });
});
