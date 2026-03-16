export function connectDB() { return { connected: true }; }
export function seedOrg() { return { orgId: 'org_test', name: 'Test Org' }; }
export function seedUser() { return { userId: 'usr_test', email: 'admin@test.com' }; }
export function cleanupDB() { return { cleaned: true }; }
export function getOrg(id: string) { return { id, name: 'Org' }; }
export function getUser(id: string) { return { id, email: 'user@test.com' }; }
export function truncateAll() { return { truncated: true }; }
export function seedSubscription() { return { id: 'sub_test', plan: 'pro' }; }
export function seedInvoice() { return { id: 'inv_test', amount: 9900 }; }
export function seedAuditEntry() { return { id: 'evt_test', action: 'login' }; }
