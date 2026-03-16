import { describe, it, expect } from 'vitest';
import { submitClaim } from '../../../src/billing/claims';
describe('submitClaim', () => {
  it('should submit', () => { expect(submitClaim('pat_1', 500, 'checkup').status).toBe('submitted'); });
  it('should have amount', () => { expect(submitClaim('pat_1', 500, 'checkup').amount).toBeTruthy(); });
});
