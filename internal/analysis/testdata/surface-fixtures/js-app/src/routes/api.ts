import { loginHandler } from '../handlers/auth.js';
import { getUsers } from '../handlers/users.js';

export function setupRoutes(app: any) {
  app.get('/api/users', getUsers);
  app.post('/api/login', loginHandler);
  app.delete('/api/users/:id', async (req: any, res: any) => {
    res.json({ deleted: true });
  });
}
