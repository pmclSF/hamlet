import { describe, it, expect } from 'vitest';
import { checkEntitlement, enforceLimit } from '../../../src/entitlements/check';

describe('checkEntitlement', () => {
  it('should return entitled', () => {
    expect(checkEntitlement('org_1', 'api_calls').entitled).toBe(true);
  });
});

describe('enforceLimit', () => {
  it('should be within limit', () => {
    expect(enforceLimit('org_1', 'api_calls', 500).withinLimit).toBe(true);
  });
});
