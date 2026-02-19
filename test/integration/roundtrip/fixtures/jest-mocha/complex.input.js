import { describe, it, expect, beforeEach, beforeAll, afterAll, jest } from '@jest/globals';
import { OrderService } from '../../src/order-service.js';
import { PaymentGateway } from '../../src/payment-gateway.js';
import { InventoryClient } from '../../src/inventory-client.js';
import { NotificationService } from '../../src/notification-service.js';

// Integration tests for the order processing pipeline
describe('OrderService', () => {
  let orderService;
  let paymentGateway;
  let notificationSpy;

  beforeAll(() => {
    process.env.ORDER_SERVICE_TIMEOUT = '5000';
  });

  afterAll(() => {
    delete process.env.ORDER_SERVICE_TIMEOUT;
  });

  beforeEach(() => {
    paymentGateway = new PaymentGateway({ testMode: true });
    orderService = new OrderService(paymentGateway);
    notificationSpy = jest.spyOn(NotificationService.prototype, 'send');
  });

  it('should create an order with a generated id', () => {
    const order = orderService.create({ productId: 'p-100', quantity: 1 });
    expect(order.id).toBeDefined();
    expect(order.status).toBe('pending');
  });

  it('should reject orders with zero quantity', () => {
    expect(() => {
      orderService.create({ productId: 'p-100', quantity: 0 });
    }).toThrow('Quantity must be at least 1');
  });

  // Payment processing tests
  describe('payment processing', () => {
    let processPaymentMock;

    beforeEach(() => {
      processPaymentMock = jest.fn().mockResolvedValue({ transactionId: 'txn-abc' });
      paymentGateway.processPayment = processPaymentMock;
    });

    it('should charge the correct amount', async () => {
      await orderService.checkout('order-1', { amount: 49.99, currency: 'USD' });
      expect(processPaymentMock).toHaveBeenCalledWith({
        amount: 49.99,
        currency: 'USD',
        orderId: 'order-1',
      });
    });

    it('should mark the order as paid on success', async () => {
      const result = await orderService.checkout('order-1', { amount: 10, currency: 'USD' });
      expect(result.status).toBe('paid');
    });

    it('should send a confirmation notification', async () => {
      await orderService.checkout('order-1', { amount: 10, currency: 'USD' });
      expect(notificationSpy).toHaveBeenCalledTimes(1);
      expect(notificationSpy).toHaveBeenCalledWith(
        expect.objectContaining({ type: 'order_confirmation' })
      );
    });

    it('should handle payment failures gracefully', async () => {
      processPaymentMock.mockRejectedValue(new Error('Card declined'));
      await expect(
        orderService.checkout('order-1', { amount: 10, currency: 'USD' })
      ).rejects.toThrow('Card declined');
    });
  });

  // Inventory reservation tests
  describe('inventory reservation', () => {
    let checkStockSpy;

    beforeEach(() => {
      checkStockSpy = jest.spyOn(InventoryClient.prototype, 'checkStock');
    });

    it('should reserve inventory for in-stock items', async () => {
      checkStockSpy.mockResolvedValue({ available: 25 });
      const reservation = await orderService.reserveInventory('p-100', 2);
      expect(reservation.reserved).toBe(true);
      expect(reservation.quantity).toBe(2);
    });

    it('should reject reservation when stock is insufficient', async () => {
      checkStockSpy.mockResolvedValue({ available: 0 });
      const reservation = await orderService.reserveInventory('p-100', 5);
      expect(reservation.reserved).toBe(false);
    });

    describe('bulk orders', () => {
      it('should apply a bulk discount for orders over 100 units', async () => {
        checkStockSpy.mockResolvedValue({ available: 500 });
        const order = await orderService.createBulkOrder('p-100', 150);
        expect(order.discount).toBeGreaterThan(0);
        expect(order.unitPrice).toBeLessThan(order.originalUnitPrice);
      });

      it('should not apply a discount for small orders', async () => {
        checkStockSpy.mockResolvedValue({ available: 500 });
        const order = await orderService.createBulkOrder('p-100', 10);
        expect(order.discount).toBe(0);
      });
    });
  });
});
