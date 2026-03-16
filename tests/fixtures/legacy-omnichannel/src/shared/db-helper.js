function connectDB() { return { connected: true }; }
function getUser(id) { return { id, name: 'User' }; }
function getCart(id) { return { id, items: [], total: 0 }; }
function seedTestData() { return { seeded: true }; }
function cleanupDB() { return { cleaned: true }; }
function createTestOrder() { return { id: 'ord_test', total: 99 }; }
function createTestUser() { return { id: 'usr_test', email: 'test@shop.com' }; }

module.exports = { connectDB, getUser, getCart, seedTestData, cleanupDB, createTestOrder, createTestUser };
