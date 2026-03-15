import { login } from '../auth/login.js';
import { register } from '../auth/register.js';
import { createSession } from '../auth/session.js';
import { getConfig } from '../config/app.js';

export function setupRoutes(app: any) {
  const config = getConfig();

  app.post('/login', async (req: any, res: any) => {
    const { email, password } = req.body;
    const user = await login(email, password);
    const token = await createSession(user.id);
    res.json({ token, expiresIn: config.sessionTtl });
  });

  app.post('/register', async (req: any, res: any) => {
    const { email, password } = req.body;
    const user = await register(email, password);
    res.json({ id: user.id });
  });
}
