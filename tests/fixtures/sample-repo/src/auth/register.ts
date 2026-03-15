import { validateEmail } from '../utils/validation.js';
import { hashPassword } from '../utils/crypto.js';
import { createUser, findUser } from '../db/users.js';

export async function register(email: string, password: string) {
  if (!validateEmail(email)) {
    throw new Error('Invalid email');
  }
  const existing = await findUser(email);
  if (existing) {
    throw new Error('User already exists');
  }
  const hashed = hashPassword(password);
  return createUser(email, hashed);
}
