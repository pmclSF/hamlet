import { validateEmail } from '../utils/validation.js';

export async function loginHandler(req: any, res: any) {
  const { email, password } = req.body;
  if (!validateEmail(email)) {
    return res.status(400).json({ error: 'Invalid email' });
  }
  res.json({ token: 'abc123' });
}

export function authMiddleware(req: any, res: any, next: any) {
  const token = req.headers.authorization;
  if (!token) {
    return res.status(401).json({ error: 'Unauthorized' });
  }
  next();
}
