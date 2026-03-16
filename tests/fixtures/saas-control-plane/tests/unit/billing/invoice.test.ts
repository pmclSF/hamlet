import { describe, it, expect } from 'vitest';
import { createInvoice } from '../../../src/billing/invoice';

describe('createInvoice', () => {
  it('should create invoice', () => {
    const r = createInvoice('org_1', 9900);
    expect(r.status).toBe('draft');
  });
  it('should reject zero amount', () => {
    expect(() => createInvoice('org_1', 0)).toThrow('Invalid amount');
  });
});
