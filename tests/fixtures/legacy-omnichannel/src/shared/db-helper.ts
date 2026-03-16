export function connectDB() { return { connected: true }; }
export function getUser(id: string) { return { id, name: 'User' }; }
export function getCart(id: string) { return { id, items: [], total: 0 }; }
export function seedTestData() { return { seeded: true }; }
export function cleanupDB() { return { cleaned: true }; }
export function createTestOrder() { return { id: 'ord_test', total: 99 }; }
export function createTestUser() { return { id: 'usr_test', email: 'test@shop.com' }; }
export function seedProducts() { return { count: 10 }; }
