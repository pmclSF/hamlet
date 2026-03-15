// HIGH FANOUT FIXTURE: imported by many test files
import { createUser, findUser, deleteUser } from '../../src/db/users.js';
import { setCache, deleteCache } from '../../src/cache/redis.js';
import { hashPassword } from '../../src/utils/crypto.js';
import { getConfig } from '../../src/config/app.js';

export async function seedDatabase() {
  const config = getConfig();
  const hash = hashPassword('password123');
  await createUser('alice@test.com', hash);
  await createUser('bob@test.com', hash);
  await createUser('carol@test.com', hash);
}

export async function cleanDatabase() {
  await deleteUser('user_1');
  await deleteUser('user_2');
  await deleteUser('user_3');
  await deleteCache('session:*');
}

export async function createTestUser(email: string = 'test@test.com') {
  const hash = hashPassword('testpass');
  return createUser(email, hash);
}

export async function findTestUser(email: string = 'test@test.com') {
  return findUser(email);
}
