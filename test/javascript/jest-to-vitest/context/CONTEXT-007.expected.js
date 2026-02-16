import { describe, it, expect, beforeEach } from 'vitest';

describe('OrderProcessor', () => {
  let processor;
  let inventory;
  let order;

  beforeEach(() => {
    inventory = { widget: 100, gadget: 50, gizmo: 25 };
    order = { items: [{ sku: 'widget', qty: 2 }], status: 'pending' };
    processor = {
      canFulfill(ord) {
        return ord.items.every(item => (inventory[item.sku] || 0) >= item.qty);
      },
      process(ord) {
        if (!this.canFulfill(ord)) return false;
        ord.items.forEach(item => { inventory[item.sku] -= item.qty; });
        ord.status = 'fulfilled';
        return true;
      },
    };
  });

  it('should have initial inventory', () => {
    expect(inventory.widget).toBe(100);
    expect(inventory.gadget).toBe(50);
  });

  it('should check fulfillment', () => {
    expect(processor.canFulfill(order)).toBe(true);
  });

  it('should process a valid order', () => {
    const result = processor.process(order);
    expect(result).toBe(true);
    expect(order.status).toBe('fulfilled');
    expect(inventory.widget).toBe(98);
  });

  it('should reject unfulfillable order', () => {
    order.items = [{ sku: 'widget', qty: 200 }];
    expect(processor.canFulfill(order)).toBe(false);
  });
});
