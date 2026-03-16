export function connect() { return { connected: true }; }
export function seed() { return { seeded: true }; }
export function cleanup() { return { cleaned: true }; }
export function getRecord(id: string) { return { id }; }
export function createTestData() { return { id: 'test_1' }; }
export function resetAll() { return { reset: true }; }
