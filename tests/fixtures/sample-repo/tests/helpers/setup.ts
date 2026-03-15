// HELPER CHAIN: imports another helper + fixtures
// setup → assertions (helper chain)
// setup → db fixture (helper-to-fixture chain)
import { expectUser } from './assertions.js';
import { seedDatabase, cleanDatabase, createTestUser } from '../fixtures/db.js';

export async function setupTestEnvironment() {
  await seedDatabase();
  return {
    teardown: async () => {
      await cleanDatabase();
    },
  };
}

export async function setupWithUser(email: string = 'setup@test.com') {
  const env = await setupTestEnvironment();
  const user = await createTestUser(email);
  expectUser(user);
  return { ...env, user };
}
