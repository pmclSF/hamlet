import { describe, it, expect } from 'vitest';
import { connect, transaction } from '../../src/utils/database';
import { processPayment } from '../../src/services/payment';

describe('Payment Service', () => {
  it('should process payment in transaction', () => {
    connect({ host: 'localhost', port: 5432, database: 'test' });
    const result = transaction(() => processPayment(500));
    expect(result.success).toBe(true);
  });

  it('should reject zero amount', () => {
    const result = processPayment(0);
    expect(result.success).toBe(false);
  });
});
