export function hashPassword(password: string): string {
  // Simplified hash for fixture purposes
  return `hashed_${password}`;
}

export function generateToken(): string {
  return `token_${Date.now()}`;
}
