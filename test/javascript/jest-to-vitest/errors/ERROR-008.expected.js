import { describe, it, expect } from 'vitest';

describe('PaymentProcessor', () => {
  it('should validate all payment fields before processing', () => {
    const payment = createPayment({
      amount: 49.99,
      currency: 'USD',
      card: '4111111111111111',
    });

    expect(payment).toBeDefined();
    expect(payment.amount).toBe(49.99);
    expect(payment.currency).toBe('USD');
    expect(payment.card).toMatch(/^\d{16}$/);
    expect(payment.status).toBe('pending');
  });

  it('should compute tax and total in sequence', () => {
    const invoice = createInvoice({ subtotal: 100, taxRate: 0.08 });
    expect(invoice.subtotal).toBe(100);
    expect(invoice.tax).toBeCloseTo(8.0);
    expect(invoice.total).toBeCloseTo(108.0);
    expect(invoice.lines).toHaveLength(1);
  });

  it('should validate the receipt fields in order', () => {
    const receipt = generateReceipt('order-123');
    expect(receipt.orderId).toBe('order-123');
    expect(receipt.timestamp).toBeDefined();
    expect(receipt.items).toBeInstanceOf(Array);
    expect(receipt.items).toHaveLength(1);
  });
});
