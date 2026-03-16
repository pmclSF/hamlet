import { describe, it, expect } from 'vitest';
import { createInvoice, finalizeInvoice, voidInvoice } from '../../../src/billing/invoice';

describe('invoice API contract', () => {
  it('createInvoice returns required fields', () => {
    const r = createInvoice('org_1', 100);
    expect(r).toHaveProperty('invoiceId');
    expect(r).toHaveProperty('amount');
    expect(r).toHaveProperty('status');
  });
  it('finalizeInvoice returns status', () => {
    expect(finalizeInvoice('inv_1')).toHaveProperty('status');
  });
});
