import { validateEmail } from '../utils/validation.js';
import { hashPassword } from '../utils/crypto.js';
import { findUser } from '../db/users.js';

export async function login(email: string, password: string) {
  if (!validateEmail(email)) {
    throw new Error('Invalid email');
  }
  const user = await findUser(email);
  if (!user) {
    throw new Error('User not found');
  }
  const hashed = hashPassword(password);
  if (hashed !== user.passwordHash) {
    throw new Error('Invalid password');
  }
  return { id: user.id, email: user.email };
}
