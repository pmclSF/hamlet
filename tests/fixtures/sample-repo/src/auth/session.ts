import { getCache, setCache } from '../cache/redis.js';

export async function createSession(userId: string): Promise<string> {
  const token = `session_${Date.now()}_${Math.random().toString(36).slice(2)}_${userId}`;
  await setCache(`session:${token}`, userId, 3600);
  return token;
}

export async function getSession(token: string): Promise<string | null> {
  return getCache(`session:${token}`);
}
