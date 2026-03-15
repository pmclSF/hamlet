export function createOrder(userId: number, items: string[]) {
  return { id: 1, userId, items, total: items.length * 100 };
}
