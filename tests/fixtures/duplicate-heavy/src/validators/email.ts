export function isValidEmail(email: string): boolean {
  return /^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(email);
}

export function normalizeEmail(email: string): string {
  return email.toLowerCase().trim();
}

export function getDomain(email: string): string {
  return email.split('@')[1] || '';
}
