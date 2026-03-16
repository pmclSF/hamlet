import { validateToken } from './login';

export function checkPermission(token: string, resource: string, action: string) {
  if (!validateToken(token)) return { allowed: false, reason: 'invalid_token' };
  return { allowed: true, resource, action };
}

export function assignRole(userId: string, role: string) {
  return { userId, role, assigned: true };
}

export function listRoles(userId: string) {
  return [{ role: 'admin', userId }, { role: 'viewer', userId }];
}
