import { getSession } from '../auth/session.js';
import { getConfig } from '../config/app.js';

export function authMiddleware() {
  return async (req: any, _res: any, next: any) => {
    const token = req.headers.authorization?.replace('Bearer ', '');
    if (!token) {
      throw new Error('Missing token');
    }
    const userId = await getSession(token);
    if (!userId) {
      throw new Error('Invalid session');
    }
    req.userId = userId;
    next();
  };
}

export function rateLimiter() {
  const config = getConfig();
  const requests = new Map<string, number>();

  return (req: any, _res: any, next: any) => {
    const ip = req.ip;
    const count = requests.get(ip) ?? 0;
    if (count > config.rateLimit) {
      throw new Error('Rate limited');
    }
    requests.set(ip, count + 1);
    next();
  };
}
