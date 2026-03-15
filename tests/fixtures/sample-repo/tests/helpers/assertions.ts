// HELPER: custom assertion utilities
// Imported by multiple test files — forms helper chain base
import { validateEmail, validatePassword } from '../../src/utils/validation.js';

export function expectValidEmail(email: string) {
  if (!validateEmail(email)) {
    throw new Error(`Expected valid email, got: ${email}`);
  }
}

export function expectValidPassword(password: string) {
  if (!validatePassword(password)) {
    throw new Error(`Expected valid password (≥8 chars), got: ${password}`);
  }
}

export function expectUser(user: any) {
  if (!user || !user.id || !user.email) {
    throw new Error(`Expected user object with id and email`);
  }
}

export function expectSession(token: string) {
  if (!token || !token.startsWith('session_')) {
    throw new Error(`Expected session token, got: ${token}`);
  }
}
