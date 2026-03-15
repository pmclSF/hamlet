// HIGH FANOUT FIXTURE: imported by auth and integration tests
import { createSession, getSession } from '../../src/auth/session.js';
import { login } from '../../src/auth/login.js';
import { register } from '../../src/auth/register.js';
import { createTestUser } from './db.js';

export async function setupAuthenticatedUser() {
  const user = await createTestUser('auth@test.com');
  const token = await createSession(user.id);
  return { user, token };
}

export async function loginAsTestUser(email: string = 'auth@test.com') {
  return login(email, 'testpass');
}

export async function registerTestUser(email: string) {
  return register(email, 'newpassword');
}

export async function getTestSession(token: string) {
  return getSession(token);
}
