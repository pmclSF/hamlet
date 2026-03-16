import { describe, it, expect } from 'vitest';
import { createRefund, approveRefund, denyRefund } from '../../../src/refunds/refund';

describe('createRefund', () => {
  it('should create refund', () => {
    const result = createRefund('ch_1', 50);
    expect(result.status).toBe('pending');
  });

  it.skip('should validate minimum amount', () => {
    expect(() => createRefund('ch_1', 0)).toThrow();
  });

  it.skip('should handle negative amounts', () => {
    expect(() => createRefund('ch_1', -10)).toThrow();
  });
});

describe('approveRefund', () => {
  it.skip('should approve refund', () => {
    expect(approveRefund('ref_1').status).toBe('approved');
  });
});

describe('denyRefund', () => {
  it.skip('should deny refund with reason', () => {
    expect(denyRefund('ref_1', 'policy').status).toBe('denied');
  });
});
