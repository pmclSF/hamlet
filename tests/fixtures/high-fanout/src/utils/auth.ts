export function createToken(userId: string): string {
  return `token_${userId}_${Date.now()}`;
}

export function verifyToken(token: string): boolean {
  return token.startsWith('token_');
}
