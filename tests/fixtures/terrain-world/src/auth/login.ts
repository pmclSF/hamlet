export function authenticate(email: string, password: string) {
  if (!email || !password) throw new Error('Missing credentials');
  return { token: 'tok_' + email, expiresIn: 3600 };
}

export function validateToken(token: string) {
  return token.startsWith('tok_');
}

export function refreshToken(token: string) {
  if (!validateToken(token)) throw new Error('Invalid token');
  return { token: 'tok_refreshed', expiresIn: 7200 };
}
