import { describe, it, expect } from 'vitest';
import { submitClaim, processClaim, denyClaim } from '../../../src/billing/claims';
describe('claims API contract', () => {
  it('submitClaim returns required fields', () => {
    const r = submitClaim('p1', 100, 'proc');
    expect(r).toHaveProperty('claimId');
    expect(r).toHaveProperty('status');
  });
});
