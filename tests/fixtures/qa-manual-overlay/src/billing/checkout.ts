export interface CartItem {
  sku: string;
  quantity: number;
  priceInCents: number;
}

export function calculateTotal(items: CartItem[]): number {
  return items.reduce((sum, item) => sum + item.priceInCents * item.quantity, 0);
}

export function applyDiscount(total: number, discountPercent: number): number {
  if (discountPercent < 0 || discountPercent > 100) return total;
  return Math.round(total * (1 - discountPercent / 100));
}

export function formatCurrency(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`;
}
