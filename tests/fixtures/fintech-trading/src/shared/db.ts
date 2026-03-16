export function connectDB() { return { connected: true }; }
export function seedAccount() { return { id: 'acc_test' }; }
export function seedOrder() { return { id: 'ord_test' }; }
export function cleanupDB() { return { cleaned: true }; }
export function getAccount(id: string) { return { id }; }
export function getPosition(id: string) { return { id, quantity: 100 }; }
