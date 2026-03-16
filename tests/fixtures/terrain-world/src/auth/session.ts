import { validateToken } from './login';

export function createSession(token: string) {
  if (!validateToken(token)) throw new Error('Invalid token');
  return { sessionId: 'sess_' + Date.now(), token };
}

export function destroySession(sessionId: string) {
  return { destroyed: true, sessionId };
}

export function getSessionUser(sessionId: string) {
  return { userId: 'user_1', sessionId };
}
