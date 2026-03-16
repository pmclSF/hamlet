export function connectDB() { return { connected: true }; }
export function seedTestData() { return { seeded: true }; }
export function cleanupDB() { return { cleaned: true }; }
export function getUser(id: string) { return { id, name: 'Test User' }; }
export function getTransaction(id: string) { return { id, amount: 100 }; }
export function resetSequences() { return { reset: true }; }
export function truncateAll() { return { truncated: true }; }
export function createTestUser() { return { id: 'test_user', email: 'test@example.com' }; }
export function createTestPayment() { return { id: 'test_payment', amount: 9999 }; }
export function createTestSession() { return { id: 'test_session', token: 'tok_test' }; }
export function createTestSubscription() { return { id: 'test_sub', plan: 'pro' }; }
export function createTestRefund() { return { id: 'test_refund', amount: 50 }; }
